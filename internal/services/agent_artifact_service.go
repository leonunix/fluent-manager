package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gorm.io/gorm"
)

type AgentArtifactInput struct {
	Name        string
	Version     string
	Description string
}

type AgentArtifactService interface {
	List() ([]models.AgentArtifact, error)
	Get(id uint) (*models.AgentArtifact, error)
	Create(input AgentArtifactInput, fileHeader *multipart.FileHeader, createdBy uint) (*models.AgentArtifact, error)
	Delete(id uint) error
}

type agentArtifactService struct {
	db         *gorm.DB
	storageDir string
}

func NewAgentArtifactService(db *gorm.DB, storageDir string) AgentArtifactService {
	return &agentArtifactService{db: db, storageDir: strings.TrimSpace(storageDir)}
}

func (s *agentArtifactService) List() ([]models.AgentArtifact, error) {
	var artifacts []models.AgentArtifact
	if err := s.db.Preload("Creator").Order("created_at DESC").Find(&artifacts).Error; err != nil {
		return nil, err
	}
	return artifacts, nil
}

func (s *agentArtifactService) Get(id uint) (*models.AgentArtifact, error) {
	var artifact models.AgentArtifact
	if err := s.db.Preload("Creator").First(&artifact, id).Error; err != nil {
		return nil, err
	}
	return &artifact, nil
}

func (s *agentArtifactService) Create(input AgentArtifactInput, fileHeader *multipart.FileHeader, createdBy uint) (*models.AgentArtifact, error) {
	if strings.TrimSpace(s.storageDir) == "" {
		return nil, fmt.Errorf("%w: agent artifact storage is not configured", ErrConflict)
	}
	if fileHeader == nil {
		return nil, fmt.Errorf("%w: file is required", ErrInvalidArgument)
	}
	if err := os.MkdirAll(s.storageDir, 0o755); err != nil {
		return nil, fmt.Errorf("create artifact storage dir: %w", err)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("open upload: %w", err)
	}
	defer src.Close()

	storageName := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), sanitizeArtifactName(fileHeader.Filename), filepath.Ext(fileHeader.Filename))
	storagePath := filepath.Join(s.storageDir, storageName)
	dst, err := os.OpenFile(storagePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("create artifact file: %w", err)
	}

	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(dst, hash), src)
	closeErr := dst.Close()
	if copyErr != nil {
		_ = os.Remove(storagePath)
		return nil, fmt.Errorf("write artifact file: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(storagePath)
		return nil, fmt.Errorf("close artifact file: %w", closeErr)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = strings.TrimSpace(fileHeader.Filename)
	}

	artifact := &models.AgentArtifact{
		Name:        name,
		Version:     strings.TrimSpace(input.Version),
		Description: strings.TrimSpace(input.Description),
		FileName:    filepath.Base(fileHeader.Filename),
		ContentType: fileHeader.Header.Get("Content-Type"),
		FileSize:    written,
		SHA256:      hex.EncodeToString(hash.Sum(nil)),
		StoragePath: storagePath,
		CreatedBy:   createdBy,
	}
	if err := s.db.Create(artifact).Error; err != nil {
		_ = os.Remove(storagePath)
		return nil, err
	}

	return s.Get(artifact.ID)
}

func (s *agentArtifactService) Delete(id uint) error {
	artifact, err := s.Get(id)
	if err != nil {
		return err
	}
	if err := s.db.Delete(&models.AgentArtifact{}, id).Error; err != nil {
		return err
	}
	_ = os.Remove(artifact.StoragePath)
	return nil
}

func sanitizeArtifactName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, base)
	base = strings.Trim(base, "-")
	if base == "" {
		return "artifact"
	}
	return base
}
