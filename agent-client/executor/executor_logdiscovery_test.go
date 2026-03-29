package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fluent-manager/fluent-manager-agent/config"
)

// newTestExecutor returns an Executor whose AllowedLogPaths is set to the given list.
func newTestExecutor(allowed []string) *Executor {
	return &Executor{cfg: &config.Config{AllowedLogPaths: allowed}}
}

// ---- isPathAllowed ----

func TestIsPathAllowed_BasicAllow(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "app.log")
	mustWriteFile(t, f, "hello")

	if !isPathAllowed(f, []string{dir}) {
		t.Fatalf("expected %q to be allowed under %q", f, dir)
	}
}

func TestIsPathAllowed_OutsidePrefix(t *testing.T) {
	allowed := t.TempDir()
	other := t.TempDir()
	f := filepath.Join(other, "secret.log")
	mustWriteFile(t, f, "secret")

	if isPathAllowed(f, []string{allowed}) {
		t.Fatalf("expected %q to be denied (outside %q)", f, allowed)
	}
}

func TestIsPathAllowed_SymlinkIntermediateDir_Escape(t *testing.T) {
	// Reproduce the reported escape:
	//   /allowed/linkdir/file.log  where linkdir → /outside
	// The resolved path is /outside/file.log, which must be denied.
	allowed := t.TempDir()
	outside := t.TempDir()

	target := filepath.Join(outside, "secret.log")
	mustWriteFile(t, target, "sensitive data")

	linkDir := filepath.Join(allowed, "linkdir")
	if err := os.Symlink(outside, linkDir); err != nil {
		t.Skipf("cannot create symlink (possibly unprivileged): %v", err)
	}

	escapedPath := filepath.Join(linkDir, "secret.log")
	if isPathAllowed(escapedPath, []string{allowed}) {
		t.Fatalf("symlink escape not caught: %q resolved outside allowed dir but was permitted", escapedPath)
	}
}

func TestIsPathAllowed_SymlinkFile_Escape(t *testing.T) {
	// /allowed/link.log → /outside/secret.log
	allowed := t.TempDir()
	outside := t.TempDir()

	target := filepath.Join(outside, "secret.log")
	mustWriteFile(t, target, "sensitive data")

	link := filepath.Join(allowed, "link.log")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	if isPathAllowed(link, []string{allowed}) {
		t.Fatalf("symlink file escape not caught: %q permitted but resolves to %q", link, target)
	}
}

func TestIsPathAllowed_NonExistentPath(t *testing.T) {
	dir := t.TempDir()
	if isPathAllowed(filepath.Join(dir, "does_not_exist.log"), []string{dir}) {
		t.Fatal("non-existent path should be denied (EvalSymlinks fails → false)")
	}
}

func TestIsPathAllowed_AllowedRootIsSymlink(t *testing.T) {
	// Reproduce: admin configures allowed_log_paths: ["/linked/var/log"]
	// where that path is itself a symlink → real directory.
	// Without resolving the root, EvalSymlinks(candidate) returns the real path,
	// which no longer matches the unresolved symlink root → file is wrongly denied.
	real := t.TempDir()
	parent := t.TempDir()
	linkRoot := filepath.Join(parent, "logs") // symlink: logs → real
	if err := os.Symlink(real, linkRoot); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	f := filepath.Join(real, "app.log")
	mustWriteFile(t, f, "line1\nline2\n")

	// Candidate path goes through the symlink root; allowed list also uses the symlink.
	candidateViaLink := filepath.Join(linkRoot, "app.log")
	if !isPathAllowed(candidateViaLink, []string{linkRoot}) {
		t.Fatalf("file under symlinked allowed root was wrongly denied")
	}
}

func TestIsPathAllowed_MultipleAllowedRoots(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	f := filepath.Join(dir2, "app.log")
	mustWriteFile(t, f, "data")

	if !isPathAllowed(f, []string{dir1, dir2}) {
		t.Fatalf("expected %q to be allowed when %q is in the list", f, dir2)
	}
}

// ---- scanLogs ----

func TestScanLogs_BasicDiscovery(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "app.log"), strings.Repeat("line\n", 20))
	mustWriteFile(t, filepath.Join(dir, "nginx.log"), strings.Repeat("req\n", 20))
	mustWriteFile(t, filepath.Join(dir, "skip.bin"), "binary")     // wrong extension
	mustWriteFile(t, filepath.Join(dir, "tiny.log"), "x")           // too small (<10 bytes)

	e := newTestExecutor([]string{dir})
	out := mustScanLogs(t, e, []string{dir})

	paths := extractPaths(out)
	if len(paths) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(paths), paths)
	}
	for _, p := range paths {
		if !strings.HasSuffix(p, ".log") {
			t.Errorf("unexpected file %q in results", p)
		}
	}
}

func TestScanLogs_DisallowedRootSkipped(t *testing.T) {
	allowed := t.TempDir()
	outside := t.TempDir()
	mustWriteFile(t, filepath.Join(outside, "app.log"), strings.Repeat("line\n", 20))

	e := newTestExecutor([]string{allowed})
	out := mustScanLogs(t, e, []string{outside})

	if len(out.Files) != 0 {
		t.Fatalf("expected 0 files when root is outside allowed list, got %d", len(out.Files))
	}
}

func TestScanLogs_SymlinkRoot(t *testing.T) {
	// Reproduce: allowed_log_paths contains a symlink directory.
	// filepath.Walk does not follow symlink dirs, so without resolving the root first
	// the scan returns empty results even though the path is in the whitelist.
	real := t.TempDir()
	parent := t.TempDir()
	linkRoot := filepath.Join(parent, "logs") // symlink: logs → real
	if err := os.Symlink(real, linkRoot); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	mustWriteFile(t, filepath.Join(real, "app.log"), strings.Repeat("line\n", 20))

	// Whitelist and scan path both use the symlink root.
	e := newTestExecutor([]string{linkRoot})
	out := mustScanLogs(t, e, []string{linkRoot})

	if len(out.Files) != 1 {
		t.Fatalf("expected 1 file via symlink root, got %d (Walk must resolve symlink before walking)", len(out.Files))
	}
}

func TestScanLogs_DefaultPathWhenNoneSpecified(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "svc.log"), strings.Repeat("log line\n", 20))

	e := newTestExecutor([]string{dir})
	// Empty paths → should default to AllowedLogPaths
	out := mustScanLogs(t, e, nil)

	if len(out.Files) != 1 {
		t.Fatalf("expected default path scan to find 1 file, got %d", len(out.Files))
	}
}

// ---- fetchLogSample ----

func TestFetchLogSample_BasicRead(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "app.log")
	mustWriteFile(t, f, "line1\nline2\nline3\n")

	e := newTestExecutor([]string{dir})
	result := mustFetchSample(t, e, []string{f}, 3)

	content, ok := result.Samples[f]
	if !ok {
		t.Fatalf("expected sample for %q", f)
	}
	if !strings.Contains(content, "line1") || !strings.Contains(content, "line3") {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestFetchLogSample_DisallowedPath(t *testing.T) {
	allowed := t.TempDir()
	outside := t.TempDir()
	f := filepath.Join(outside, "secret.log")
	mustWriteFile(t, f, strings.Repeat("secret\n", 20))

	e := newTestExecutor([]string{allowed})
	result := mustFetchSample(t, e, []string{f}, 10)

	if got := result.Samples[f]; !strings.Contains(got, "[error:") {
		t.Fatalf("expected error for disallowed path, got %q", got)
	}
}

func TestFetchLogSample_SymlinkIntermediateDirEscape(t *testing.T) {
	allowed := t.TempDir()
	outside := t.TempDir()
	target := filepath.Join(outside, "secret.log")
	mustWriteFile(t, target, strings.Repeat("secret\n", 20))

	linkDir := filepath.Join(allowed, "linkdir")
	if err := os.Symlink(outside, linkDir); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	escapedPath := filepath.Join(linkDir, "secret.log")

	e := newTestExecutor([]string{allowed})
	result := mustFetchSample(t, e, []string{escapedPath}, 10)

	if got := result.Samples[escapedPath]; !strings.Contains(got, "[error:") {
		t.Fatalf("symlink escape not caught in fetchLogSample: got %q", got)
	}
}

func TestFetchLogSample_FileCapEnforced(t *testing.T) {
	dir := t.TempDir()
	var paths []string
	for i := 0; i < 15; i++ {
		f := filepath.Join(dir, fmt.Sprintf("app%d.log", i))
		mustWriteFile(t, f, strings.Repeat("data\n", 20))
		paths = append(paths, f)
	}

	e := newTestExecutor([]string{dir})
	result := mustFetchSample(t, e, paths, 10)

	// Agent caps at 10 files; at most 10 entries (errors count too, but we expect real results here)
	if len(result.Samples) > 10 {
		t.Fatalf("expected at most 10 samples, got %d", len(result.Samples))
	}
}

func TestFetchLogSample_PerFileByteCap(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "big.log")
	// Write > maxBytesPerFile (25 KB) of log lines
	line := strings.Repeat("x", 100) + "\n"
	mustWriteFile(t, f, strings.Repeat(line, 500)) // 500 × 101 bytes ≈ 50 KB

	e := newTestExecutor([]string{dir})
	result := mustFetchSample(t, e, []string{f}, 500)

	content := result.Samples[f]
	if len(content) > maxBytesPerFile {
		t.Fatalf("per-file content exceeds cap: %d bytes", len(content))
	}
	// Must end on a complete line (no partial line at end)
	if content != "" && !strings.HasSuffix(content, "x") {
		t.Errorf("content does not end at a line boundary: tail=%q", content[max(0, len(content)-20):])
	}
}

func TestFetchLogSample_TotalBudgetCap(t *testing.T) {
	dir := t.TempDir()
	// 10 files × ~30 KB each → triggers the total budget (250 KB) after ~8 files.
	line := strings.Repeat("a", 100) + "\n"
	var paths []string
	for i := 0; i < 10; i++ {
		f := filepath.Join(dir, fmt.Sprintf("svc%d.log", i))
		mustWriteFile(t, f, strings.Repeat(line, 300)) // ~30 KB each
		paths = append(paths, f)
	}

	e := newTestExecutor([]string{dir})
	result := mustFetchSample(t, e, paths, 500)

	total := 0
	for _, v := range result.Samples {
		total += len(v)
	}
	if total > maxTotalBytes {
		t.Fatalf("total sample bytes %d exceeds budget %d", total, maxTotalBytes)
	}
}

func TestFetchLogSample_OutputIsValidJSON(t *testing.T) {
	dir := t.TempDir()
	line := strings.Repeat("z", 100) + "\n"
	var paths []string
	for i := 0; i < 10; i++ {
		f := filepath.Join(dir, fmt.Sprintf("f%d.log", i))
		mustWriteFile(t, f, strings.Repeat(line, 450)) // near 45 KB each
		paths = append(paths, f)
	}

	e := newTestExecutor([]string{dir})
	rawArgs, _ := json.Marshal(fetchLogSampleArgs{Files: paths, Lines: 500})
	out, err := e.fetchLogSample(string(rawArgs))
	if err != nil {
		t.Fatalf("fetchLogSample returned error: %v", err)
	}

	var parsed fetchLogSampleResult
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\n(first 200 bytes: %q)", err, out[:min(200, len(out))])
	}
}

func TestFetchLogSample_OutputFitsDBLimit_QuoteHeavy(t *testing.T) {
	// Regression test for: 400 KB raw quote content → ~820 KB JSON (2× expansion).
	// Verifies that the content budgets (maxBytesPerFile, maxTotalBytes) are sized
	// conservatively enough that the marshaled output always fits in maxOutputBytes.
	dir := t.TempDir()
	// Each line is all double-quotes — worst-case JSON expansion (each " → \")
	line := strings.Repeat(`"`, 100) + "\n"
	var paths []string
	for i := 0; i < 10; i++ {
		f := filepath.Join(dir, fmt.Sprintf("q%d.log", i))
		mustWriteFile(t, f, strings.Repeat(line, 500)) // 500 × 101 bytes ≈ 50 KB each
		paths = append(paths, f)
	}

	e := newTestExecutor([]string{dir})
	rawArgs, _ := json.Marshal(fetchLogSampleArgs{Files: paths, Lines: 500})
	out, err := e.fetchLogSample(string(rawArgs))
	if err != nil {
		t.Fatalf("fetchLogSample returned error: %v", err)
	}

	// Must be valid JSON
	var parsed fetchLogSampleResult
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON with quote-heavy content: %v", err)
	}

	// Must fit in the DB column limit
	if len(out) > maxOutputBytes {
		t.Fatalf("JSON output %d bytes exceeds DB limit %d (raw budget invariant broken)", len(out), maxOutputBytes)
	}
}

func TestFetchLogSample_OutputFitsDBLimit_NulHeavy(t *testing.T) {
	// Regression: NUL bytes (\x00) expand to \u0000 (6 bytes each) when JSON-encoded,
	// which can inflate 250 KB of raw content to ~1.5 MB — well beyond the 512 KB DB limit.
	// After sanitization (NUL → space), expansion stays ≤2× and the output must fit.
	dir := t.TempDir()
	line := strings.Repeat("\x00", 100) + "\n"
	var paths []string
	for i := 0; i < 10; i++ {
		f := filepath.Join(dir, fmt.Sprintf("n%d.log", i))
		mustWriteFile(t, f, strings.Repeat(line, 500)) // ~50 KB of NUL bytes each
		paths = append(paths, f)
	}

	e := newTestExecutor([]string{dir})
	rawArgs, _ := json.Marshal(fetchLogSampleArgs{Files: paths, Lines: 500})
	out, err := e.fetchLogSample(string(rawArgs))
	if err != nil {
		t.Fatalf("fetchLogSample returned error: %v", err)
	}

	var parsed fetchLogSampleResult
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON with NUL content: %v", err)
	}
	if len(out) > maxOutputBytes {
		t.Fatalf("JSON output %d bytes exceeds DB limit %d (NUL sanitization missing)", len(out), maxOutputBytes)
	}
}

// ---- tailFile ----

func TestTailFile_ExactLines(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "t.log")
	mustWriteFile(t, f, "a\nb\nc\nd\ne\n")
	lines, err := tailFile(f, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "c" || lines[2] != "e" {
		t.Errorf("wrong lines: %v", lines)
	}
}

func TestTailFile_FewerLinesThanN(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "t.log")
	mustWriteFile(t, f, "only\ntwo\n")
	lines, err := tailFile(f, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d", len(lines))
	}
}

func TestTailFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "empty.log")
	mustWriteFile(t, f, "")
	lines, err := tailFile(f, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 0 {
		t.Fatalf("expected 0 lines for empty file, got %d", len(lines))
	}
}

func TestTailFile_ByteCapPreventsOOM(t *testing.T) {
	// File with no newlines: should not read more than maxCollect bytes
	dir := t.TempDir()
	f := filepath.Join(dir, "noeol.log")
	mustWriteFile(t, f, strings.Repeat("X", 600*1024)) // 600 KB, no newlines
	lines, err := tailFile(f, 10)
	if err != nil {
		t.Fatal(err)
	}
	// Should return at most 1 "line" (the whole capped chunk, no \n found)
	total := 0
	for _, l := range lines {
		total += len(l)
	}
	if total > int(maxCollect)+1024 {
		t.Fatalf("tailFile read %d bytes from a 600 KB no-newline file (expected ≤ %d)", total, maxCollect)
	}
}

// ---- helpers ----

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func mustScanLogs(t *testing.T, e *Executor, paths []string) scanLogsResult {
	t.Helper()
	args := scanLogsArgs{Paths: paths}
	raw, _ := json.Marshal(args)
	out, err := e.scanLogs(string(raw))
	if err != nil {
		t.Fatalf("scanLogs error: %v", err)
	}
	var result scanLogsResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("scanLogs output is not valid JSON: %v", err)
	}
	return result
}

func mustFetchSample(t *testing.T, e *Executor, files []string, lines int) fetchLogSampleResult {
	t.Helper()
	args := fetchLogSampleArgs{Files: files, Lines: lines}
	raw, _ := json.Marshal(args)
	out, err := e.fetchLogSample(string(raw))
	if err != nil {
		t.Fatalf("fetchLogSample error: %v", err)
	}
	var result fetchLogSampleResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("fetchLogSample output is not valid JSON: %v", err)
	}
	return result
}

func extractPaths(r scanLogsResult) []string {
	paths := make([]string, len(r.Files))
	for i, f := range r.Files {
		paths[i] = f.Path
	}
	return paths
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
