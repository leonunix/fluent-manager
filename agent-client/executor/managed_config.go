package executor

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const managedFluentBitParserFilename = "fluent-manager-parsers.conf"

var fluentBitSectionHeaderRe = regexp.MustCompile(`^\s*\[([A-Za-z_]+)\]\s*$`)

type preparedConfig struct {
	mainContent          string
	managedParserPath    string
	managedParserContent string
	sourceHash           string
}

type configBackupMetadata struct {
	ManagedParserPath       string `json:"managed_parser_path,omitempty"`
	ManagedParserBackupPath string `json:"managed_parser_backup_path,omitempty"`
	ManagedParserExisted    bool   `json:"managed_parser_existed"`
	SourceHash              string `json:"source_hash,omitempty"`
	SourceHashSet           bool   `json:"source_hash_set"`
}

type fluentBitChunk struct {
	kind    string
	lines   []string
	isBlock bool
}

func hashConfigContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", sum)
}

func normalizeTrailingNewline(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}
	return strings.TrimRight(content, "\n") + "\n"
}

func (e *Executor) prepareManagedConfig(content string) preparedConfig {
	prepared := preparedConfig{
		mainContent: normalizeTrailingNewline(content),
		sourceHash:  hashConfigContent(content),
	}
	if e.cfg.FluentType != "fluentbit" {
		return prepared
	}

	parserPath := e.managedFluentBitParserPath()
	if parserPath == "" {
		return prepared
	}

	parserFiles := e.fluentBitDefaultParserFiles()
	mainContent, parserContent := rewriteFluentBitConfigForManagedParsers(content, parserPath, parserFiles)
	prepared.mainContent = normalizeTrailingNewline(mainContent)
	prepared.managedParserPath = parserPath
	prepared.managedParserContent = normalizeTrailingNewline(parserContent)
	return prepared
}

func (e *Executor) writePreparedConfig(prepared preparedConfig) error {
	if prepared.managedParserPath != "" {
		if prepared.managedParserContent == "" {
			if err := os.Remove(prepared.managedParserPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove managed parser file: %w", err)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(prepared.managedParserPath), 0o755); err != nil {
				return fmt.Errorf("create managed parser dir: %w", err)
			}
			if err := os.WriteFile(prepared.managedParserPath, []byte(prepared.managedParserContent), 0o644); err != nil {
				return fmt.Errorf("write managed parser file: %w", err)
			}
		}
	}

	if err := os.WriteFile(e.cfg.FluentConfigPath, []byte(prepared.mainContent), 0o644); err != nil {
		return fmt.Errorf("write main config: %w", err)
	}
	return nil
}

func (e *Executor) managedFluentBitParserPath() string {
	baseDir := strings.TrimSpace(e.cfg.FluentConfigDir)
	if baseDir == "" {
		baseDir = filepath.Dir(strings.TrimSpace(e.cfg.FluentConfigPath))
	}
	if baseDir == "" || baseDir == "." {
		return ""
	}
	return filepath.Join(baseDir, managedFluentBitParserFilename)
}

func (e *Executor) fluentBitDefaultParserFiles() []string {
	candidates := []string{}
	configPath := strings.TrimSpace(e.cfg.FluentConfigPath)
	if configPath != "" {
		baseDir := filepath.Dir(configPath)
		candidates = append(candidates,
			filepath.Join(baseDir, "parsers.conf"),
			filepath.Join(baseDir, "parsers_multiline.conf"),
		)
	}
	candidates = append(candidates,
		"/etc/fluent-bit/parsers.conf",
		"/etc/fluent-bit/parsers_multiline.conf",
	)

	files := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}
		if fileExists(candidate) {
			files = append(files, candidate)
		}
	}
	return files
}

func (e *Executor) appliedConfigHashPath() string {
	stateDir := strings.TrimSpace(e.cfg.StateDir)
	if stateDir == "" {
		return ""
	}
	return filepath.Join(stateDir, ".applied_config_hash")
}

func (e *Executor) readAppliedConfigHash() string {
	path := e.appliedConfigHashPath()
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (e *Executor) writeAppliedConfigHash(hash string) error {
	path := e.appliedConfigHashPath()
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create hash state dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(strings.TrimSpace(hash)), 0o600); err != nil {
		return fmt.Errorf("write applied config hash: %w", err)
	}
	return nil
}

func (e *Executor) removeAppliedConfigHash() error {
	path := e.appliedConfigHashPath()
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove applied config hash: %w", err)
	}
	return nil
}

func (e *Executor) writeBackupMetadata(backupPath string) error {
	meta := configBackupMetadata{}
	if parserPath := e.managedFluentBitParserPath(); parserPath != "" {
		meta.ManagedParserPath = parserPath
		if fileExists(parserPath) {
			parserBackupPath := backupPath + ".parser"
			data, err := os.ReadFile(parserPath)
			if err != nil {
				return fmt.Errorf("read managed parser file: %w", err)
			}
			if err := os.WriteFile(parserBackupPath, data, 0o644); err != nil {
				return fmt.Errorf("write parser backup: %w", err)
			}
			meta.ManagedParserExisted = true
			meta.ManagedParserBackupPath = parserBackupPath
		}
	}

	if hash := e.readAppliedConfigHash(); hash != "" {
		meta.SourceHash = hash
		meta.SourceHashSet = true
	}

	if meta.ManagedParserPath == "" && !meta.SourceHashSet {
		return nil
	}

	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal backup metadata: %w", err)
	}
	if err := os.WriteFile(backupPath+".meta.json", data, 0o644); err != nil {
		return fmt.Errorf("write backup metadata: %w", err)
	}
	return nil
}

func (e *Executor) restoreBackupMetadata(backupPath string) {
	data, err := os.ReadFile(backupPath + ".meta.json")
	if err != nil {
		return
	}

	var meta configBackupMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		log.Printf("[executor] restore metadata decode failed: %v", err)
		return
	}

	if meta.ManagedParserPath != "" {
		if meta.ManagedParserExisted {
			parserData, err := os.ReadFile(meta.ManagedParserBackupPath)
			if err != nil {
				log.Printf("[executor] restore managed parser backup failed: %v", err)
			} else {
				if err := os.MkdirAll(filepath.Dir(meta.ManagedParserPath), 0o755); err != nil {
					log.Printf("[executor] create managed parser dir during rollback failed: %v", err)
				} else if err := os.WriteFile(meta.ManagedParserPath, parserData, 0o644); err != nil {
					log.Printf("[executor] restore managed parser file failed: %v", err)
				}
			}
		} else if err := os.Remove(meta.ManagedParserPath); err != nil && !os.IsNotExist(err) {
			log.Printf("[executor] remove managed parser file during rollback failed: %v", err)
		}
	}

	if meta.SourceHashSet {
		if err := e.writeAppliedConfigHash(meta.SourceHash); err != nil {
			log.Printf("[executor] restore applied config hash failed: %v", err)
		}
	} else if err := e.removeAppliedConfigHash(); err != nil {
		log.Printf("[executor] clear applied config hash failed: %v", err)
	}
}

func removeBackupArtifacts(backupPath string) {
	for _, candidate := range []string{
		backupPath,
		backupPath + ".meta.json",
		backupPath + ".parser",
	} {
		if err := os.Remove(candidate); err != nil && !os.IsNotExist(err) {
			log.Printf("[executor] remove backup artifact %s failed: %v", candidate, err)
		}
	}
}

func splitFluentBitChunks(content string) []fluentBitChunk {
	lines := strings.Split(content, "\n")
	chunks := make([]fluentBitChunk, 0, 8)
	cursor := 0

	if len(lines) == 0 {
		return chunks
	}

	for cursor < len(lines) {
		matches := fluentBitSectionHeaderRe.FindStringSubmatch(strings.TrimSpace(lines[cursor]))
		if len(matches) == 2 {
			break
		}
		cursor++
	}
	if cursor > 0 {
		chunks = append(chunks, fluentBitChunk{
			lines: append([]string(nil), lines[:cursor]...),
		})
	}

	for cursor < len(lines) {
		matches := fluentBitSectionHeaderRe.FindStringSubmatch(strings.TrimSpace(lines[cursor]))
		if len(matches) != 2 {
			start := cursor
			for cursor < len(lines) {
				if len(fluentBitSectionHeaderRe.FindStringSubmatch(strings.TrimSpace(lines[cursor]))) == 2 {
					break
				}
				cursor++
			}
			chunks = append(chunks, fluentBitChunk{
				lines: append([]string(nil), lines[start:cursor]...),
			})
			continue
		}

		start := cursor
		kind := strings.ToUpper(matches[1])
		cursor++
		for cursor < len(lines) {
			if len(fluentBitSectionHeaderRe.FindStringSubmatch(strings.TrimSpace(lines[cursor]))) == 2 {
				break
			}
			cursor++
		}
		chunks = append(chunks, fluentBitChunk{
			kind:    kind,
			lines:   append([]string(nil), lines[start:cursor]...),
			isBlock: true,
		})
	}

	return chunks
}

func rewriteFluentBitConfigForManagedParsers(content, managedParserPath string, defaultParserFiles []string) (string, string) {
	chunks := splitFluentBitChunks(content)
	if len(chunks) == 0 {
		return strings.TrimRight(content, "\n"), ""
	}

	preamble := []string{}
	startIdx := 0
	if !chunks[0].isBlock {
		preamble = append([]string(nil), chunks[0].lines...)
		startIdx = 1
	}

	parserSections := make([]string, 0, 4)
	nonParserChunks := make([]fluentBitChunk, 0, len(chunks))
	for idx := startIdx; idx < len(chunks); idx++ {
		chunk := chunks[idx]
		if chunk.isBlock && (chunk.kind == "PARSER" || chunk.kind == "MULTILINE_PARSER") {
			block := strings.TrimSpace(strings.Join(chunk.lines, "\n"))
			if block != "" {
				parserSections = append(parserSections, block)
			}
			continue
		}
		nonParserChunks = append(nonParserChunks, chunk)
	}

	requiredParserFiles := make([]string, 0, len(defaultParserFiles)+1)
	if len(parserSections) > 0 {
		requiredParserFiles = append(requiredParserFiles, managedParserPath)
	}
	if fluentBitConfigUsesParserReference(nonParserChunks) {
		requiredParserFiles = append(requiredParserFiles, defaultParserFiles...)
	}
	requiredParserFiles = uniqueNonEmptyStrings(requiredParserFiles)

	bodyLines := make([]string, 0, len(nonParserChunks)*4)
	serviceHandled := false
	for _, chunk := range nonParserChunks {
		if chunk.isBlock && chunk.kind == "SERVICE" {
			bodyLines = append(bodyLines, rewriteFluentBitServiceBlock(chunk.lines, managedParserPath, requiredParserFiles)...)
			serviceHandled = true
			continue
		}
		bodyLines = append(bodyLines, chunk.lines...)
	}

	outLines := append([]string{}, preamble...)
	if len(requiredParserFiles) > 0 && !serviceHandled {
		if len(outLines) > 0 && strings.TrimSpace(outLines[len(outLines)-1]) != "" {
			outLines = append(outLines, "")
		}
		outLines = append(outLines, "[SERVICE]")
		for _, parserFile := range requiredParserFiles {
			outLines = append(outLines, "    Parsers_File "+parserFile)
		}
		outLines = append(outLines, "")
	}
	outLines = append(outLines, bodyLines...)

	return strings.TrimRight(strings.Join(outLines, "\n"), "\n"), strings.Join(parserSections, "\n\n")
}

func rewriteFluentBitServiceBlock(lines []string, managedParserPath string, requiredParserFiles []string) []string {
	filtered := make([]string, 0, len(lines)+1)
	existingParserFiles := make(map[string]struct{}, len(requiredParserFiles))
	for idx, line := range lines {
		if idx > 0 && isManagedParsersFileLine(line, managedParserPath) {
			continue
		}
		if idx > 0 {
			if parserFile, ok := fluentBitParserFileValue(line); ok {
				existingParserFiles[parserFile] = struct{}{}
			}
		}
		filtered = append(filtered, line)
	}

	if len(requiredParserFiles) == 0 {
		return filtered
	}

	insertAt := len(filtered)
	for insertAt > 1 && strings.TrimSpace(filtered[insertAt-1]) == "" {
		insertAt--
	}

	out := append([]string{}, filtered[:insertAt]...)
	for _, parserFile := range requiredParserFiles {
		if _, exists := existingParserFiles[parserFile]; exists {
			continue
		}
		out = append(out, "    Parsers_File "+parserFile)
	}
	out = append(out, filtered[insertAt:]...)
	return out
}

func isManagedParsersFileLine(line, parserPath string) bool {
	if parserPath == "" {
		return false
	}
	parserFile, ok := fluentBitParserFileValue(line)
	return ok && parserFile == parserPath
}

func fluentBitParserFileValue(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
		return "", false
	}
	fields := strings.Fields(trimmed)
	if len(fields) < 2 || !strings.EqualFold(fields[0], "parsers_file") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(trimmed, fields[0])), true
}

func fluentBitConfigUsesParserReference(chunks []fluentBitChunk) bool {
	for _, chunk := range chunks {
		if !chunk.isBlock {
			continue
		}
		for _, line := range chunk.lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
				continue
			}
			fields := strings.Fields(trimmed)
			if len(fields) < 2 {
				continue
			}
			key := strings.ToLower(strings.TrimSpace(fields[0]))
			if isFluentBitParserReferenceKey(key) {
				return true
			}
		}
	}
	return false
}

func isFluentBitParserReferenceKey(key string) bool {
	switch key {
	case "parser", "parser_firstline", "parser_n", "multiline.parser":
		return true
	default:
		return strings.HasPrefix(key, "parser_")
	}
}

func uniqueNonEmptyStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
