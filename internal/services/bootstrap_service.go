package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fluent-manager/fluent-manager/internal/models"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

const (
	bootstrapTaskStatusPending   = "pending"
	bootstrapTaskStatusRunning   = "running"
	bootstrapTaskStatusCompleted = "completed"
	bootstrapTaskStatusFailed    = "failed"
	maxBootstrapHostBatchSize    = 500
)

type BootstrapSettings struct {
	DefaultAgentAPIKey     string
	Secret                 string
	DisableHostKeyChecking bool
}

type BootstrapCapability struct {
	Supported               bool     `json:"supported"`
	AnsiblePlaybookPath     string   `json:"ansible_playbook_path"`
	SSHPassPath             string   `json:"sshpass_path"`
	RolePath                string   `json:"role_path"`
	DefaultAgentBinaryPath  string   `json:"default_agent_binary_path"`
	DefaultAgentBinaryFound bool     `json:"default_agent_binary_found"`
	Reasons                 []string `json:"reasons"`
}

type BootstrapHostInput struct {
	Hostname       string `json:"hostname"`
	IPAddress      string `json:"ip_address"`
	SSHPort        int    `json:"ssh_port"`
	SSHUser        string `json:"ssh_user"`
	AuthType       string `json:"auth_type"`
	Password       string `json:"password"`
	PrivateKey     string `json:"private_key"`
	BecomePassword string `json:"become_password"`
	NodeUID        string `json:"node_uid"`
	Labels         string `json:"labels"`
	ClusterID      *uint  `json:"cluster_id"`
	Description    string `json:"description"`
}

type BootstrapHostFilters struct {
	ClusterID     string `json:"cluster_id"`
	EnvironmentID string `json:"environment_id"`
	DataCenterID  string `json:"datacenter_id"`
	RegionID      string `json:"region_id"`
	AuthType      string `json:"auth_type"`
	Search        string `json:"search"`
}

type BootstrapTaskInput struct {
	Name               string               `json:"name"`
	ServerURL          string               `json:"server_url"`
	AgentAPIKey        string               `json:"agent_api_key"`
	AgentAccessKeyID   *uint                `json:"agent_access_key_id"`
	ClusterID          *uint                `json:"cluster_id"`
	FluentType       string               `json:"fluent_type"`
	InstallRuntime   bool                 `json:"install_runtime"`
	AgentBinaryPath  string               `json:"agent_binary_path"`
	AgentDownloadURL string               `json:"agent_download_url"`
	AllMatching      bool                 `json:"all_matching"`
	Filters          BootstrapHostFilters `json:"filters"`
	HostIDs          []uint               `json:"host_ids"`
	Hosts            []BootstrapHostInput `json:"hosts"`
}

type BootstrapService interface {
	GetCapability() BootstrapCapability
	ListHosts(filters BootstrapHostFilters, allowedClusters []uint, page, pageSize int) ([]models.BootstrapHost, int64, error)
	CreateHost(input BootstrapHostInput, createdBy uint, allowedClusters []uint) (*models.BootstrapHost, error)
	CreateHosts(inputs []BootstrapHostInput, createdBy uint, allowedClusters []uint) ([]models.BootstrapHost, error)
	UpdateHost(id uint, input BootstrapHostInput, allowedClusters []uint) (*models.BootstrapHost, error)
	DeleteHost(id uint, allowedClusters []uint) error
	Create(input BootstrapTaskInput, userID uint, allowedClusters []uint) (*models.BootstrapTask, error)
	List(page, pageSize int, allowedClusters []uint) ([]models.BootstrapTask, int64, error)
	Get(id uint, page, pageSize int, allowedClusters []uint) (*models.BootstrapTask, []models.BootstrapRecord, int64, error)
	Close()
}

type bootstrapService struct {
	db              *gorm.DB
	settings        BootstrapSettings
	cipher          *bootstrapSecretCipher
	cipherInitError string
	agentKeysSvc    AgentAccessKeyService
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewBootstrapService(db *gorm.DB, settings BootstrapSettings, agentKeysSvc AgentAccessKeyService) BootstrapService {
	ctx, cancel := context.WithCancel(context.Background())
	var cipher *bootstrapSecretCipher
	secret := strings.TrimSpace(settings.Secret)
	if secret == "" {
		log.Printf("WARNING: bootstrap credential encryption unavailable because bootstrap secret is empty")
	} else {
		if c, err := newBootstrapSecretCipher(secret); err == nil {
			cipher = c
		} else {
			log.Printf("WARNING: bootstrap credential encryption initialization failed: %v", err)
			return &bootstrapService{
				db:              db,
				settings:        settings,
				cipherInitError: err.Error(),
				agentKeysSvc:    agentKeysSvc,
				ctx:             ctx,
				cancel:          cancel,
			}
		}
	}
	return &bootstrapService{db: db, settings: settings, cipher: cipher, agentKeysSvc: agentKeysSvc, ctx: ctx, cancel: cancel}
}

func (s *bootstrapService) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *bootstrapService) GetCapability() BootstrapCapability {
	capability := BootstrapCapability{}

	if ansiblePath, err := exec.LookPath("ansible-playbook"); err == nil {
		capability.AnsiblePlaybookPath = ansiblePath
	} else {
		capability.Reasons = append(capability.Reasons, "ansible-playbook is not installed or not in PATH")
	}

	if sshpassPath, err := exec.LookPath("sshpass"); err == nil {
		capability.SSHPassPath = sshpassPath
	}

	rolePath, binaryPath := findAnsibleAssets()
	if rolePath != "" {
		capability.RolePath = rolePath
	} else {
		capability.Reasons = append(capability.Reasons, "Ansible role directory scripts/ansible/roles/fluent_manager_agent was not found")
	}

	if binaryPath != "" {
		capability.DefaultAgentBinaryPath = binaryPath
		capability.DefaultAgentBinaryFound = true
	}

	if s.cipher == nil {
		reason := "bootstrap credential encryption is unavailable because the server secret is missing"
		if strings.TrimSpace(s.cipherInitError) != "" {
			reason = "bootstrap credential encryption is unavailable: " + s.cipherInitError
		}
		capability.Reasons = append(capability.Reasons, reason)
	}

	capability.Supported = capability.AnsiblePlaybookPath != "" && capability.RolePath != "" && s.cipher != nil
	return capability
}

func (s *bootstrapService) ListHosts(filters BootstrapHostFilters, allowedClusters []uint, page, pageSize int) ([]models.BootstrapHost, int64, error) {
	var hosts []models.BootstrapHost
	query := s.db.Preload("Cluster.Region.DataCenter").Preload("Creator")
	query = applyBootstrapHostFilters(query, filters, allowedClusters)

	var total int64
	query.Model(&models.BootstrapHost{}).Count(&total)
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&hosts).Error; err != nil {
		return nil, 0, err
	}
	return hosts, total, nil
}

func (s *bootstrapService) CreateHost(input BootstrapHostInput, createdBy uint, allowedClusters []uint) (*models.BootstrapHost, error) {
	resolved, err := s.prepareStoredHost(nil, input, true, allowedClusters)
	if err != nil {
		return nil, err
	}
	resolved.CreatedBy = createdBy
	if err := s.db.Create(resolved).Error; err != nil {
		return nil, err
	}
	return s.reloadHost(resolved.ID, allowedClusters)
}

func (s *bootstrapService) CreateHosts(inputs []BootstrapHostInput, createdBy uint, allowedClusters []uint) ([]models.BootstrapHost, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("%w: at least one host is required", ErrInvalidArgument)
	}
	if len(inputs) > maxBootstrapHostBatchSize {
		return nil, fmt.Errorf("%w: a single request can include at most %d hosts", ErrInvalidArgument, maxBootstrapHostBatchSize)
	}

	preparedHosts := make([]*models.BootstrapHost, 0, len(inputs))
	for _, input := range inputs {
		prepared, err := s.prepareStoredHost(nil, input, true, allowedClusters)
		if err != nil {
			return nil, err
		}
		prepared.CreatedBy = createdBy
		preparedHosts = append(preparedHosts, prepared)
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, host := range preparedHosts {
			if err := tx.Create(host).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	created := make([]models.BootstrapHost, 0, len(preparedHosts))
	for _, host := range preparedHosts {
		reloaded, err := s.reloadHost(host.ID, allowedClusters)
		if err != nil {
			return nil, err
		}
		created = append(created, *reloaded)
	}
	return created, nil
}

func (s *bootstrapService) UpdateHost(id uint, input BootstrapHostInput, allowedClusters []uint) (*models.BootstrapHost, error) {
	current, err := s.getScopedHost(id, allowedClusters)
	if err != nil {
		return nil, err
	}
	updated, err := s.prepareStoredHost(current, input, false, allowedClusters)
	if err != nil {
		return nil, err
	}
	if err := s.db.Model(current).Select(bootstrapHostUpdatableColumns()).Updates(updated).Error; err != nil {
		return nil, err
	}
	return s.reloadHost(id, allowedClusters)
}

func (s *bootstrapService) DeleteHost(id uint, allowedClusters []uint) error {
	host, err := s.getScopedHost(id, allowedClusters)
	if err != nil {
		return err
	}
	return s.db.Delete(&models.BootstrapHost{}, host.ID).Error
}

func (s *bootstrapService) Create(input BootstrapTaskInput, userID uint, allowedClusters []uint) (*models.BootstrapTask, error) {
	resolved, records, err := s.validateAndPrepareTask(input, allowedClusters)
	if err != nil {
		return nil, err
	}

	task := &models.BootstrapTask{
		Name:              resolved.Name,
		Status:            bootstrapTaskStatusPending,
		Message:           "Task queued. Waiting for Ansible execution.",
		ClusterID:         resolved.ClusterID,
		FluentType:        resolved.FluentType,
		InstallRuntime:    resolved.InstallRuntime,
		ServerURL:         resolved.ServerURL,
		AgentBinarySource: resolved.AgentBinarySource,
		AgentBinaryPath:   resolved.AgentBinaryPath,
		AgentDownloadURL:  resolved.AgentDownloadURL,
		TotalHosts:        len(records),
		CreatedBy:         userID,
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		for idx := range records {
			records[idx].BootstrapTaskID = task.ID
		}
		if err := tx.Create(&records).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := s.db.Preload("Cluster.Region.DataCenter").Preload("Creator").First(task, task.ID).Error; err != nil {
		return nil, err
	}

	go s.runTask(task.ID, resolved)
	return task, nil
}

func (s *bootstrapService) List(page, pageSize int, allowedClusters []uint) ([]models.BootstrapTask, int64, error) {
	query := s.db.Preload("Cluster.Region.DataCenter").Preload("Creator")
	if isScopedRequest(allowedClusters) {
		if len(allowedClusters) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where("id IN (SELECT DISTINCT bootstrap_task_id FROM bootstrap_records WHERE cluster_id IN ?)", allowedClusters)
		}
	}

	var total int64
	query.Model(&models.BootstrapTask{}).Count(&total)

	var tasks []models.BootstrapTask
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

func (s *bootstrapService) Get(id uint, page, pageSize int, allowedClusters []uint) (*models.BootstrapTask, []models.BootstrapRecord, int64, error) {
	var task models.BootstrapTask
	query := s.db.Preload("Cluster.Region.DataCenter").Preload("Creator")
	if err := query.First(&task, id).Error; err != nil {
		return nil, nil, 0, err
	}

	if isScopedRequest(allowedClusters) {
		if len(allowedClusters) == 0 {
			return nil, nil, 0, gorm.ErrRecordNotFound
		}
		var count int64
		if err := s.db.Model(&models.BootstrapRecord{}).
			Where("bootstrap_task_id = ? AND cluster_id IN ?", id, allowedClusters).
			Count(&count).Error; err != nil {
			return nil, nil, 0, err
		}
		if count == 0 {
			return nil, nil, 0, gorm.ErrRecordNotFound
		}
	}

	countQuery := s.db.Model(&models.BootstrapRecord{}).Where("bootstrap_task_id = ?", id)
	countQuery = applyAllowedClusterScope(countQuery, "cluster_id", allowedClusters)

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, nil, 0, err
	}

	recordQuery := s.db.Where("bootstrap_task_id = ?", id)
	recordQuery = applyAllowedClusterScope(recordQuery, "cluster_id", allowedClusters)

	var records []models.BootstrapRecord
	if err := recordQuery.
		Preload("Node.Cluster.Region.DataCenter").
		Preload("BootstrapHost.Cluster.Region.DataCenter").
		Preload("Cluster.Region.DataCenter").
		Order("id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, nil, 0, err
	}
	return &task, records, total, nil
}

type preparedBootstrapTask struct {
	Name              string
	ServerURL         string
	AgentAPIKey       string
	ClusterID         *uint
	FluentType        string
	InstallRuntime    bool
	AgentBinarySource string
	AgentBinaryPath   string
	AgentDownloadURL  string
	Hosts             []preparedBootstrapHost
}

type preparedBootstrapHost struct {
	Alias           string
	BootstrapHostID *uint
	ClusterID       *uint
	Hostname        string
	IPAddress       string
	SSHPort         int
	SSHUser         string
	AuthType        string
	Password        string
	PrivateKey      string
	BecomePassword  string
	NodeUID         string
	Labels          string
}

func (s *bootstrapService) prepareStoredHost(existing *models.BootstrapHost, input BootstrapHostInput, creating bool, allowedClusters []uint) (*models.BootstrapHost, error) {
	if s.cipher == nil {
		return nil, fmt.Errorf("%w: bootstrap credential encryption is unavailable", ErrConflict)
	}

	authType := strings.ToLower(strings.TrimSpace(input.AuthType))
	if authType == "" {
		if !creating && existing != nil {
			authType = existing.AuthType
		} else {
			authType = "private_key"
		}
	}
	input.AuthType = authType

	requireCredential := creating || existing == nil || existing.AuthType != authType
	if !requireCredential {
		switch authType {
		case "password":
			requireCredential = !existing.HasPassword
		case "private_key":
			requireCredential = !existing.HasPrivateKey
		}
	}

	preparedHost, _, err := prepareBootstrapHost(0, input, true, s.GetCapability().SSHPassPath != "", requireCredential)
	if err != nil {
		return nil, err
	}

	clusterID, err := s.resolveScopedClusterID(input.ClusterID, allowedClusters)
	if err != nil {
		return nil, err
	}
	if isScopedRequest(allowedClusters) && clusterID == nil {
		return nil, fmt.Errorf("%w: scoped users must bind saved hosts to an allowed cluster", ErrForbidden)
	}

	host := &models.BootstrapHost{
		Hostname:    preparedHost.Hostname,
		IPAddress:   preparedHost.IPAddress,
		SSHPort:     preparedHost.SSHPort,
		SSHUser:     preparedHost.SSHUser,
		AuthType:    preparedHost.AuthType,
		NodeUID:     preparedHost.NodeUID,
		Labels:      preparedHost.Labels,
		ClusterID:   clusterID,
		Description: strings.TrimSpace(input.Description),
	}

	if existing != nil {
		host.CreatedBy = existing.CreatedBy
	}

	switch preparedHost.AuthType {
	case "password":
		password := strings.TrimSpace(input.Password)
		if !creating && password == "" {
			if existing == nil || !existing.HasPassword {
				return nil, fmt.Errorf("%w: password is required for password auth", ErrInvalidArgument)
			}
			host.PasswordEncrypted = existing.PasswordEncrypted
			host.HasPassword = true
		} else {
			encrypted, err := s.cipher.Encrypt(password)
			if err != nil {
				return nil, err
			}
			host.PasswordEncrypted = encrypted
			host.HasPassword = password != ""
		}
		host.PrivateKeyEncrypted = ""
		host.HasPrivateKey = false
	case "private_key":
		privateKey := strings.TrimSpace(input.PrivateKey)
		if !creating && privateKey == "" {
			if existing == nil || !existing.HasPrivateKey {
				return nil, fmt.Errorf("%w: private_key is required for private key auth", ErrInvalidArgument)
			}
			host.PrivateKeyEncrypted = existing.PrivateKeyEncrypted
			host.HasPrivateKey = true
		} else {
			encrypted, err := s.cipher.Encrypt(privateKey)
			if err != nil {
				return nil, err
			}
			host.PrivateKeyEncrypted = encrypted
			host.HasPrivateKey = privateKey != ""
		}
		host.PasswordEncrypted = ""
		host.HasPassword = false
	}

	becomePassword := strings.TrimSpace(input.BecomePassword)
	if !creating && becomePassword == "" && existing != nil && existing.HasBecomePassword {
		host.BecomePasswordEncrypted = existing.BecomePasswordEncrypted
		host.HasBecomePassword = true
	} else if becomePassword != "" {
		encrypted, err := s.cipher.Encrypt(becomePassword)
		if err != nil {
			return nil, err
		}
		host.BecomePasswordEncrypted = encrypted
		host.HasBecomePassword = true
	} else {
		host.BecomePasswordEncrypted = ""
		host.HasBecomePassword = false
	}

	return host, nil
}

func (s *bootstrapService) validateAndPrepareTask(input BootstrapTaskInput, allowedClusters []uint) (*preparedBootstrapTask, []models.BootstrapRecord, error) {
	capability := s.GetCapability()
	if !capability.Supported {
		return nil, nil, fmt.Errorf("%w: %s", ErrConflict, strings.Join(capability.Reasons, "; "))
	}

	serverURL := strings.TrimSpace(input.ServerURL)
	if serverURL == "" {
		return nil, nil, fmt.Errorf("%w: server_url is required", ErrInvalidArgument)
	}

	fluentType := strings.ToLower(strings.TrimSpace(input.FluentType))
	if fluentType == "" {
		fluentType = "fluentbit"
	}
	switch fluentType {
	case "fluentbit", "fluentd", "auto":
	default:
		return nil, nil, fmt.Errorf("%w: fluent_type must be one of fluentbit, fluentd, auto", ErrInvalidArgument)
	}

	clusterID, err := s.resolveScopedClusterID(input.ClusterID, allowedClusters)
	if err != nil {
		return nil, nil, err
	}

	agentAPIKey := strings.TrimSpace(input.AgentAPIKey)
	if agentAPIKey == "" && input.AgentAccessKeyID != nil && s.agentKeysSvc != nil {
		resolved, err := s.agentKeysSvc.GetPlaintextKey(*input.AgentAccessKeyID)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: failed to resolve agent_access_key_id: %v", ErrInvalidArgument, err)
		}
		agentAPIKey = resolved
	}
	if agentAPIKey == "" {
		agentAPIKey = strings.TrimSpace(s.settings.DefaultAgentAPIKey)
	}
	if agentAPIKey == "" {
		return nil, nil, fmt.Errorf("%w: agent_api_key is required because the server has no default agent.api_key configured", ErrInvalidArgument)
	}

	agentBinaryPath, agentDownloadURL, agentBinarySource, err := validateAgentBinaryInput(input.AgentBinaryPath, input.AgentDownloadURL, capability)
	if err != nil {
		return nil, nil, err
	}

	inlineHosts, inlineRecords, err := s.prepareInlineHosts(input.Hosts, capability.SSHPassPath != "")
	if err != nil {
		return nil, nil, err
	}

	savedHosts, savedRecords, err := s.prepareSavedHosts(input.HostIDs, input.Filters, input.AllMatching, clusterID, allowedClusters)
	if err != nil {
		return nil, nil, err
	}

	hosts := append(savedHosts, inlineHosts...)
	records := append(savedRecords, inlineRecords...)
	if len(hosts) == 0 {
		return nil, nil, fmt.Errorf("%w: at least one host or host_id is required", ErrInvalidArgument)
	}
	if len(hosts) > maxBootstrapHostBatchSize {
		return nil, nil, fmt.Errorf("%w: a single bootstrap task can target at most %d hosts", ErrInvalidArgument, maxBootstrapHostBatchSize)
	}
	if isScopedRequest(allowedClusters) {
		for _, record := range records {
			if record.ClusterID == nil {
				return nil, nil, fmt.Errorf("%w: scoped users must assign each target host to an allowed cluster or set a task cluster override", ErrForbidden)
			}
		}
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = fmt.Sprintf("Bootstrap %s (%d hosts)", strings.ToUpper(fluentType), len(hosts))
	}

	return &preparedBootstrapTask{
		Name:              name,
		ServerURL:         serverURL,
		AgentAPIKey:       agentAPIKey,
		ClusterID:         clusterID,
		FluentType:        fluentType,
		InstallRuntime:    input.InstallRuntime,
		AgentBinarySource: agentBinarySource,
		AgentBinaryPath:   agentBinaryPath,
		AgentDownloadURL:  agentDownloadURL,
		Hosts:             hosts,
	}, records, nil
}

func validateAgentBinaryInput(agentBinaryPathRaw, agentDownloadURLRaw string, capability BootstrapCapability) (string, string, string, error) {
	agentBinaryPath := strings.TrimSpace(agentBinaryPathRaw)
	agentDownloadURL := strings.TrimSpace(agentDownloadURLRaw)
	switch {
	case agentBinaryPath != "":
		absPath, err := filepath.Abs(agentBinaryPath)
		if err != nil {
			return "", "", "", fmt.Errorf("%w: invalid agent_binary_path", ErrInvalidArgument)
		}
		if _, err := os.Stat(absPath); err != nil {
			return "", "", "", fmt.Errorf("%w: agent binary not found at %s", ErrInvalidArgument, absPath)
		}
		return absPath, "", "local_path", nil
	case agentDownloadURL != "":
		return "", agentDownloadURL, "download_url", nil
	case capability.DefaultAgentBinaryFound:
		return capability.DefaultAgentBinaryPath, "", "local_path", nil
	default:
		return "", "", "", fmt.Errorf("%w: provide agent_binary_path or agent_download_url, or place the binary at %s", ErrInvalidArgument, defaultAgentBinaryHint())
	}
}

func (s *bootstrapService) prepareInlineHosts(inputs []BootstrapHostInput, sshpassAvailable bool) ([]preparedBootstrapHost, []models.BootstrapRecord, error) {
	if len(inputs) == 0 {
		return nil, nil, nil
	}
	hosts := make([]preparedBootstrapHost, 0, len(inputs))
	records := make([]models.BootstrapRecord, 0, len(inputs))
	for idx, input := range inputs {
		prepared, record, err := prepareBootstrapHost(idx, input, true, sshpassAvailable, true)
		if err != nil {
			return nil, nil, err
		}
		hosts = append(hosts, prepared)
		records = append(records, record)
	}
	return hosts, records, nil
}

func (s *bootstrapService) prepareSavedHosts(hostIDs []uint, filters BootstrapHostFilters, allMatching bool, taskClusterID *uint, allowedClusters []uint) ([]preparedBootstrapHost, []models.BootstrapRecord, error) {
	storedHosts, err := s.loadScopedHostsForTask(hostIDs, filters, allMatching, allowedClusters)
	if err != nil {
		return nil, nil, err
	}
	if len(storedHosts) == 0 {
		return nil, nil, nil
	}
	hosts := make([]preparedBootstrapHost, 0, len(storedHosts))
	records := make([]models.BootstrapRecord, 0, len(storedHosts))
	for idx, stored := range storedHosts {

		password, err := s.cipher.Decrypt(stored.PasswordEncrypted)
		if err != nil {
			return nil, nil, err
		}
		privateKey, err := s.cipher.Decrypt(stored.PrivateKeyEncrypted)
		if err != nil {
			return nil, nil, err
		}
		becomePassword, err := s.cipher.Decrypt(stored.BecomePasswordEncrypted)
		if err != nil {
			return nil, nil, err
		}

		clusterID := stored.ClusterID
		if taskClusterID != nil {
			clusterID = taskClusterID
		}

		prepared := preparedBootstrapHost{
			Alias:           fmt.Sprintf("saved_host_%d", idx+1),
			BootstrapHostID: &stored.ID,
			ClusterID:       clusterID,
			Hostname:        stored.Hostname,
			IPAddress:       stored.IPAddress,
			SSHPort:         stored.SSHPort,
			SSHUser:         stored.SSHUser,
			AuthType:        stored.AuthType,
			Password:        password,
			PrivateKey:      privateKey,
			BecomePassword:  becomePassword,
			NodeUID:         stored.NodeUID,
			Labels:          stored.Labels,
		}
		record := models.BootstrapRecord{
			BootstrapHostID: &stored.ID,
			Hostname:        stored.Hostname,
			IPAddress:       stored.IPAddress,
			SSHPort:         stored.SSHPort,
			SSHUser:         stored.SSHUser,
			AuthType:        stored.AuthType,
			NodeUID:         stored.NodeUID,
			Labels:          stored.Labels,
			ClusterID:       clusterID,
			Alias:           prepared.Alias,
			Status:          bootstrapTaskStatusPending,
			Message:         "Waiting for Ansible execution.",
		}
		hosts = append(hosts, prepared)
		records = append(records, record)
	}
	return hosts, records, nil
}

func (s *bootstrapService) loadScopedHostsForTask(hostIDs []uint, filters BootstrapHostFilters, allMatching bool, allowedClusters []uint) ([]models.BootstrapHost, error) {
	seen := map[uint]struct{}{}
	result := make([]models.BootstrapHost, 0)
	appendUnique := func(items []models.BootstrapHost) {
		for _, item := range items {
			if _, ok := seen[item.ID]; ok {
				continue
			}
			seen[item.ID] = struct{}{}
			result = append(result, item)
		}
	}

	uniqueIDs := uniqueUintSlice(hostIDs)
	if len(uniqueIDs) > 0 {
		query := s.db.Preload("Cluster.Region.DataCenter").Preload("Creator")
		query = applyAllowedClusterScope(query, "bootstrap_hosts.cluster_id", allowedClusters)
		var explicit []models.BootstrapHost
		if err := query.Where("bootstrap_hosts.id IN ?", uniqueIDs).Find(&explicit).Error; err != nil {
			return nil, err
		}
		if len(explicit) != len(uniqueIDs) {
			return nil, fmt.Errorf("%w: some selected hosts were not found or are outside your scope", ErrForbidden)
		}
		appendUnique(explicit)
	}

	if allMatching {
		query := s.db.Preload("Cluster.Region.DataCenter").Preload("Creator")
		query = applyBootstrapHostFilters(query, filters, allowedClusters)
		var matched []models.BootstrapHost
		if err := query.Find(&matched).Error; err != nil {
			return nil, err
		}
		appendUnique(matched)
	}

	return result, nil
}

func prepareBootstrapHost(idx int, host BootstrapHostInput, generateAlias bool, sshpassAvailable bool, requireCredential bool) (preparedBootstrapHost, models.BootstrapRecord, error) {
	hostname := strings.TrimSpace(host.Hostname)
	if hostname == "" {
		return preparedBootstrapHost{}, models.BootstrapRecord{}, fmt.Errorf("%w: hostname is required for host #%d", ErrInvalidArgument, idx+1)
	}
	ipAddress := strings.TrimSpace(host.IPAddress)
	if ipAddress != "" && net.ParseIP(ipAddress) == nil {
		return preparedBootstrapHost{}, models.BootstrapRecord{}, fmt.Errorf("%w: invalid ip_address for host %s", ErrInvalidArgument, hostname)
	}
	sshUser := strings.TrimSpace(host.SSHUser)
	if sshUser == "" {
		return preparedBootstrapHost{}, models.BootstrapRecord{}, fmt.Errorf("%w: ssh_user is required for host %s", ErrInvalidArgument, hostname)
	}

	sshPort := host.SSHPort
	if sshPort == 0 {
		sshPort = 22
	}
	if sshPort < 1 || sshPort > 65535 {
		return preparedBootstrapHost{}, models.BootstrapRecord{}, fmt.Errorf("%w: invalid ssh_port for host %s", ErrInvalidArgument, hostname)
	}

	authType := strings.ToLower(strings.TrimSpace(host.AuthType))
	if authType == "" {
		authType = "private_key"
	}
	switch authType {
	case "password":
		if requireCredential && strings.TrimSpace(host.Password) == "" {
			return preparedBootstrapHost{}, models.BootstrapRecord{}, fmt.Errorf("%w: password is required for host %s", ErrInvalidArgument, hostname)
		}
		if !sshpassAvailable {
			return preparedBootstrapHost{}, models.BootstrapRecord{}, fmt.Errorf("%w: sshpass is required on the server for password-based SSH bootstrap", ErrConflict)
		}
	case "private_key":
		if requireCredential && strings.TrimSpace(host.PrivateKey) == "" {
			return preparedBootstrapHost{}, models.BootstrapRecord{}, fmt.Errorf("%w: private_key is required for host %s", ErrInvalidArgument, hostname)
		}
	default:
		return preparedBootstrapHost{}, models.BootstrapRecord{}, fmt.Errorf("%w: auth_type must be password or private_key", ErrInvalidArgument)
	}

	clusterID := host.ClusterID
	alias := ""
	if generateAlias {
		alias = fmt.Sprintf("bootstrap_%d", idx+1)
	}
	prepared := preparedBootstrapHost{
		Alias:          alias,
		ClusterID:      clusterID,
		Hostname:       hostname,
		IPAddress:      ipAddress,
		SSHPort:        sshPort,
		SSHUser:        sshUser,
		AuthType:       authType,
		Password:       strings.TrimSpace(host.Password),
		PrivateKey:     strings.TrimSpace(host.PrivateKey),
		BecomePassword: strings.TrimSpace(host.BecomePassword),
		NodeUID:        strings.TrimSpace(host.NodeUID),
		Labels:         strings.TrimSpace(host.Labels),
	}
	record := models.BootstrapRecord{
		Hostname:  hostname,
		IPAddress: ipAddress,
		SSHPort:   sshPort,
		SSHUser:   sshUser,
		AuthType:  authType,
		NodeUID:   prepared.NodeUID,
		Labels:    prepared.Labels,
		ClusterID: clusterID,
		Alias:     alias,
		Status:    bootstrapTaskStatusPending,
		Message:   "Waiting for Ansible execution.",
	}
	return prepared, record, nil
}

func (s *bootstrapService) resolveScopedClusterID(clusterID *uint, allowedClusters []uint) (*uint, error) {
	if clusterID == nil {
		return nil, nil
	}
	if isScopedRequest(allowedClusters) && !containsUint(allowedClusters, *clusterID) {
		return nil, fmt.Errorf("%w: cluster is outside current scope", ErrForbidden)
	}
	var cluster models.Cluster
	if err := s.db.First(&cluster, *clusterID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: cluster not found", ErrInvalidArgument)
		}
		return nil, err
	}
	return clusterID, nil
}

func (s *bootstrapService) getScopedHost(id uint, allowedClusters []uint) (*models.BootstrapHost, error) {
	var host models.BootstrapHost
	query := s.db.Preload("Cluster.Region.DataCenter").Preload("Creator")
	query = applyAllowedClusterScope(query, "cluster_id", allowedClusters)
	if err := query.First(&host, id).Error; err != nil {
		return nil, err
	}
	return &host, nil
}

func (s *bootstrapService) reloadHost(id uint, allowedClusters []uint) (*models.BootstrapHost, error) {
	return s.getScopedHost(id, allowedClusters)
}

func (s *bootstrapService) runTask(taskID uint, input *preparedBootstrapTask) {
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	now := time.Now()
	if err := s.updateTask(taskID, map[string]interface{}{
		"status":     bootstrapTaskStatusRunning,
		"message":    "Running ansible-playbook against target hosts.",
		"started_at": &now,
	}); err != nil {
		s.logTaskError(taskID, "failed to mark bootstrap task as running", err)
		return
	}
	if err := s.db.Model(&models.BootstrapRecord{}).
		Where("bootstrap_task_id = ?", taskID).
		Updates(map[string]interface{}{
			"status":  bootstrapTaskStatusRunning,
			"message": "Ansible playbook is running.",
		}).Error; err != nil {
		s.logTaskError(taskID, "failed to mark bootstrap records as running", err)
	}

	workDir, cleanup, err := prepareBootstrapWorkspace(s.GetCapability(), input, s.settings.DisableHostKeyChecking)
	if err != nil {
		s.failWholeTask(taskID, err)
		return
	}
	defer cleanup()

	stdout, stderr, runErr := executeBootstrapPlaybook(ctx, workDir, input.AgentAPIKey, s.settings.DisableHostKeyChecking)
	recap := parseAnsibleRecap(stdout + "\n" + stderr)
	hostErrors := parseHostErrors(stdout + "\n" + stderr)

	var records []models.BootstrapRecord
	if err := s.db.Where("bootstrap_task_id = ?", taskID).Find(&records).Error; err != nil {
		s.failWholeTask(taskID, err)
		return
	}

	successCount := 0
	failCount := 0
	if len(recap) == 0 && runErr != nil {
		for _, record := range records {
			failCount++
			if err := s.updateRecord(record.ID, map[string]interface{}{
				"status":         bootstrapTaskStatusFailed,
				"message":        summarizeOutput(stderr, stdout, runErr.Error()),
				"output_excerpt": truncateString(stdout+"\n"+stderr, 4000),
			}); err != nil {
				s.failWholeTask(taskID, fmt.Errorf("failed to persist bootstrap record result: %w", err))
				return
			}
		}
		if err := s.finishTask(taskID, successCount, failCount, fmt.Sprintf("Ansible execution failed before a per-host recap was produced: %s", summarizeOutput(stderr, stdout, runErr.Error()))); err != nil {
			s.logTaskError(taskID, "failed to persist final bootstrap task status", err)
		}
		return
	}

	registrationCheckCancelled := false
	for _, record := range records {
		recapItem, ok := recap[record.Alias]
		if !ok {
			failCount++
			if err := s.updateRecord(record.ID, map[string]interface{}{
				"status":         bootstrapTaskStatusFailed,
				"message":        "Host did not appear in the Ansible recap. Inspect task output for details.",
				"output_excerpt": truncateString(stdout+"\n"+stderr, 4000),
			}); err != nil {
				s.failWholeTask(taskID, fmt.Errorf("failed to persist bootstrap record result: %w", err))
				return
			}
			continue
		}

		if recapItem.Unreachable > 0 || recapItem.Failed > 0 {
			failCount++
			message := hostErrors[record.Alias]
			if message == "" {
				message = recapItem.String()
			}
			if err := s.updateRecord(record.ID, map[string]interface{}{
				"status":         bootstrapTaskStatusFailed,
				"message":        message,
				"output_excerpt": truncateString(stdout+"\n"+stderr, 4000),
			}); err != nil {
				s.failWholeTask(taskID, fmt.Errorf("failed to persist bootstrap record result: %w", err))
				return
			}
			continue
		}

		successCount++
		nodeMessage := "Installation finished. Waiting for the first agent heartbeat."
		var nodeID *uint
		if !registrationCheckCancelled {
			var cancelled bool
			nodeID, nodeMessage, cancelled = s.waitForNodeRegistration(ctx, record.ClusterID, record.Hostname, record.IPAddress, record.NodeUID)
			if cancelled {
				registrationCheckCancelled = true
				nodeMessage = "Installation finished. Registration check stopped because the server is shutting down."
			}
		} else {
			nodeMessage = "Installation finished. Registration check skipped because the server is shutting down."
		}
		updates := map[string]interface{}{
			"status":         bootstrapTaskStatusCompleted,
			"message":        nodeMessage,
			"output_excerpt": recapItem.String(),
		}
		if nodeID != nil {
			updates["node_id"] = *nodeID
		}
		if err := s.updateRecord(record.ID, updates); err != nil {
			s.failWholeTask(taskID, fmt.Errorf("failed to persist bootstrap record result: %w", err))
			return
		}
	}

	message := "Ansible bootstrap finished."
	if failCount > 0 {
		message = fmt.Sprintf("Bootstrap finished with %d success and %d failed hosts.", successCount, failCount)
	} else {
		message = fmt.Sprintf("Bootstrap finished successfully for %d hosts.", successCount)
	}
	if err := s.finishTask(taskID, successCount, failCount, message); err != nil {
		s.logTaskError(taskID, "failed to persist final bootstrap task status", err)
	}
}

func (s *bootstrapService) waitForNodeRegistration(ctx context.Context, clusterID *uint, hostname, ipAddress, nodeUID string) (*uint, string, bool) {
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return nil, "Installation finished. Registration check stopped because the server is shutting down.", true
		}

		var node models.Node
		query := s.db.Model(&models.Node{})
		switch {
		case strings.TrimSpace(nodeUID) != "":
			query = query.Where("node_uid = ?", nodeUID)
		case strings.TrimSpace(ipAddress) != "":
			query = query.Where("hostname = ? OR ip_address = ?", hostname, ipAddress)
		default:
			query = query.Where("hostname = ?", hostname)
		}
		err := query.First(&node).Error
		if err == nil {
			message := fmt.Sprintf("Installation finished and agent registered as node %s.", node.Hostname)
			if clusterID != nil && node.ClusterID == nil {
				message = fmt.Sprintf("%s Cluster assignment was left unchanged because node ownership must follow the normal registration flow.", message)
			}
			return &node.ID, message, false
		}
		timer := time.NewTimer(3 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, "Installation finished. Registration check stopped because the server is shutting down.", true
		case <-timer.C:
		}
	}
	return nil, "Installation finished. Waiting for the first agent heartbeat.", false
}

func (s *bootstrapService) finishTask(taskID uint, successCount, failCount int, message string) error {
	status := bootstrapTaskStatusCompleted
	if failCount > 0 {
		status = bootstrapTaskStatusFailed
	}
	now := time.Now()
	return s.updateTask(taskID, map[string]interface{}{
		"status":        status,
		"message":       message,
		"success_count": successCount,
		"fail_count":    failCount,
		"finished_at":   &now,
	})
}

func (s *bootstrapService) failWholeTask(taskID uint, err error) {
	var records []models.BootstrapRecord
	if dbErr := s.db.Where("bootstrap_task_id = ?", taskID).Find(&records).Error; dbErr != nil {
		s.logTaskError(taskID, "failed to load bootstrap task records while marking task failed", dbErr)
		records = nil
	}
	for _, record := range records {
		if dbErr := s.updateRecord(record.ID, map[string]interface{}{
			"status":  bootstrapTaskStatusFailed,
			"message": truncateString(err.Error(), 2000),
		}); dbErr != nil {
			s.logTaskError(taskID, fmt.Sprintf("failed to mark bootstrap record %d as failed", record.ID), dbErr)
		}
	}
	if dbErr := s.finishTask(taskID, 0, len(records), truncateString(err.Error(), 2000)); dbErr != nil {
		s.logTaskError(taskID, "failed to mark bootstrap task as failed", dbErr)
	}
}

type inventoryFile struct {
	All bootstrapInventoryGroup `yaml:"all"`
}

type bootstrapInventoryGroup struct {
	Hosts map[string]map[string]interface{} `yaml:"hosts"`
}

type playbookPlay struct {
	Name        string                   `yaml:"name"`
	Hosts       string                   `yaml:"hosts"`
	Become      bool                     `yaml:"become"`
	GatherFacts bool                     `yaml:"gather_facts"`
	Vars        map[string]interface{}   `yaml:"vars"`
	Roles       []map[string]interface{} `yaml:"roles"`
}

func prepareBootstrapWorkspace(capability BootstrapCapability, input *preparedBootstrapTask, disableHostKeyChecking bool) (string, func(), error) {
	if !capability.Supported {
		return "", func() {}, fmt.Errorf("bootstrap capability unavailable: %s", strings.Join(capability.Reasons, "; "))
	}

	workDir, err := os.MkdirTemp("", "fm-bootstrap-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(workDir) }

	inventory := inventoryFile{
		All: bootstrapInventoryGroup{
			Hosts: map[string]map[string]interface{}{},
		},
	}

	for _, host := range input.Hosts {
		ansibleHost := host.IPAddress
		if strings.TrimSpace(ansibleHost) == "" {
			ansibleHost = host.Hostname
		}
		hostVars := map[string]interface{}{
			"ansible_host": ansibleHost,
			"ansible_user": host.SSHUser,
			"ansible_port": host.SSHPort,
		}
		if disableHostKeyChecking {
			hostVars["ansible_ssh_common_args"] = "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
		}
		if strings.TrimSpace(host.BecomePassword) != "" {
			hostVars["ansible_become_password"] = host.BecomePassword
		}
		if host.AuthType == "password" {
			hostVars["ansible_password"] = host.Password
		} else {
			keyPath, err := writePrivateKey(workDir, host.Alias, host.PrivateKey)
			if err != nil {
				cleanup()
				return "", func() {}, err
			}
			hostVars["ansible_ssh_private_key_file"] = keyPath
		}
		if host.NodeUID != "" {
			hostVars["fm_node_uid"] = host.NodeUID
		}
		if host.Labels != "" {
			hostVars["fm_labels"] = host.Labels
		}
		inventory.All.Hosts[host.Alias] = hostVars
	}

	inventoryBytes, err := yaml.Marshal(inventory)
	if err != nil {
		cleanup()
		return "", func() {}, err
	}
	if err := os.WriteFile(filepath.Join(workDir, "inventory.yml"), inventoryBytes, 0o600); err != nil {
		cleanup()
		return "", func() {}, err
	}

	playbook := []playbookPlay{
		{
			Name:        "Bootstrap Fluent Manager agent",
			Hosts:       "all",
			Become:      true,
			GatherFacts: true,
			Vars: map[string]interface{}{
				"fm_server_url":         input.ServerURL,
				"fm_api_key":            "{{ lookup('env', 'FM_AGENT_API_KEY') }}",
				"fm_fluent_type":        input.FluentType,
				"fm_install_fluentbit":  input.InstallRuntime && input.FluentType == "fluentbit",
				"fm_install_fluentd":    input.InstallRuntime && input.FluentType == "fluentd",
				"fm_agent_binary":       input.AgentBinaryPath,
				"fm_agent_download_url": input.AgentDownloadURL,
			},
			Roles: []map[string]interface{}{
				{"role": capability.RolePath},
			},
		},
	}

	playbookBytes, err := yaml.Marshal(playbook)
	if err != nil {
		cleanup()
		return "", func() {}, err
	}
	if err := os.WriteFile(filepath.Join(workDir, "playbook.yml"), playbookBytes, 0o600); err != nil {
		cleanup()
		return "", func() {}, err
	}

	ansibleCfgLines := []string{
		"[defaults]",
		"retry_files_enabled = False",
		"interpreter_python = auto_silent",
	}
	if disableHostKeyChecking {
		ansibleCfgLines = append(ansibleCfgLines, "host_key_checking = False")
	}
	ansibleCfg := strings.Join(ansibleCfgLines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(workDir, "ansible.cfg"), []byte(ansibleCfg), 0o600); err != nil {
		cleanup()
		return "", func() {}, err
	}

	return workDir, cleanup, nil
}

func writePrivateKey(workDir, alias, privateKey string) (string, error) {
	keyPath := filepath.Join(workDir, alias+".pem")
	content := strings.TrimSpace(privateKey)
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if err := os.WriteFile(keyPath, []byte(content), 0o600); err != nil {
		return "", err
	}
	return keyPath, nil
}

func executeBootstrapPlaybook(parent context.Context, workDir, agentAPIKey string, disableHostKeyChecking bool) (string, string, error) {
	ansiblePath, err := exec.LookPath("ansible-playbook")
	if err != nil {
		return "", "", err
	}

	ctx, cancel := context.WithTimeout(parent, 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, ansiblePath, "-i", "inventory.yml", "playbook.yml")
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(),
		"ANSIBLE_CONFIG="+filepath.Join(workDir, "ansible.cfg"),
		"FM_AGENT_API_KEY="+agentAPIKey,
	)
	if disableHostKeyChecking {
		cmd.Env = append(cmd.Env, "ANSIBLE_HOST_KEY_CHECKING=False")
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		err = fmt.Errorf("ansible execution timed out after 30 minutes")
	}
	return stdoutBuf.String(), stderrBuf.String(), err
}

type ansibleRecap struct {
	OK          int
	Changed     int
	Unreachable int
	Failed      int
}

func (r ansibleRecap) String() string {
	return fmt.Sprintf("ok=%d changed=%d unreachable=%d failed=%d", r.OK, r.Changed, r.Unreachable, r.Failed)
}

var (
	recapLinePattern   = regexp.MustCompile(`(?m)^([A-Za-z0-9_.-]+)\s*:\s*ok=(\d+)\s+changed=(\d+)\s+unreachable=(\d+)\s+failed=(\d+)`)
	hostFailurePattern = regexp.MustCompile(`(?m)^(fatal|changed): \[([A-Za-z0-9_.-]+)\]: (FAILED|UNREACHABLE)! => (.+)$`)
)

func parseAnsibleRecap(output string) map[string]ansibleRecap {
	matches := recapLinePattern.FindAllStringSubmatch(output, -1)
	result := make(map[string]ansibleRecap, len(matches))
	for _, match := range matches {
		if len(match) != 6 {
			continue
		}
		var item ansibleRecap
		fmt.Sscanf(match[2], "%d", &item.OK)
		fmt.Sscanf(match[3], "%d", &item.Changed)
		fmt.Sscanf(match[4], "%d", &item.Unreachable)
		fmt.Sscanf(match[5], "%d", &item.Failed)
		result[match[1]] = item
	}
	return result
}

func parseHostErrors(output string) map[string]string {
	matches := hostFailurePattern.FindAllStringSubmatch(output, -1)
	result := map[string]string{}
	for _, match := range matches {
		if len(match) != 5 {
			continue
		}
		result[match[2]] = truncateString(match[4], 1000)
	}
	return result
}

func summarizeOutput(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return truncateString(value, 1000)
		}
	}
	return "ansible execution failed"
}

func truncateString(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || value == "" {
		return ""
	}
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit]) + "..."
}

func bootstrapHostUpdatableColumns() []string {
	return []string{
		"hostname",
		"ip_address",
		"ssh_port",
		"ssh_user",
		"auth_type",
		"password_encrypted",
		"private_key_encrypted",
		"become_password_encrypted",
		"has_password",
		"has_private_key",
		"has_become_password",
		"node_uid",
		"labels",
		"description",
		"cluster_id",
		"created_by",
	}
}

func isScopedRequest(allowedClusters []uint) bool {
	return allowedClusters != nil
}

func applyAllowedClusterScope(query *gorm.DB, field string, allowedClusters []uint) *gorm.DB {
	if !isScopedRequest(allowedClusters) {
		return query
	}
	if len(allowedClusters) == 0 {
		return query.Where("1 = 0")
	}
	return query.Where(field+" IN ?", allowedClusters)
}

func applyBootstrapHostFilters(query *gorm.DB, filters BootstrapHostFilters, allowedClusters []uint) *gorm.DB {
	query = applyAllowedClusterScope(query, "bootstrap_hosts.cluster_id", allowedClusters)
	if filters.ClusterID != "" {
		query = query.Where("bootstrap_hosts.cluster_id = ?", filters.ClusterID)
	}
	if filters.EnvironmentID != "" {
		query = query.Where("bootstrap_hosts.cluster_id IN (SELECT id FROM clusters WHERE environment_id = ?)", filters.EnvironmentID)
	}
	if filters.DataCenterID != "" {
		query = query.Joins("JOIN clusters bhc2 ON bhc2.id = bootstrap_hosts.cluster_id").
			Joins("JOIN regions bhr2 ON bhr2.id = bhc2.region_id").
			Where("bhr2.data_center_id = ?", filters.DataCenterID)
	}
	if filters.RegionID != "" {
		query = query.Joins("JOIN clusters bhc3 ON bhc3.id = bootstrap_hosts.cluster_id").
			Where("bhc3.region_id = ?", filters.RegionID)
	}
	if filters.AuthType != "" {
		query = query.Where("bootstrap_hosts.auth_type = ?", filters.AuthType)
	}
	if filters.Search != "" {
		term := "%" + filters.Search + "%"
		query = query.Where(
			"bootstrap_hosts.hostname LIKE ? OR bootstrap_hosts.ip_address LIKE ? OR bootstrap_hosts.node_uid LIKE ? OR bootstrap_hosts.description LIKE ? OR bootstrap_hosts.labels LIKE ?",
			term, term, term, term, term,
		)
	}
	return query
}

func (s *bootstrapService) updateTask(taskID uint, updates map[string]interface{}) error {
	return s.db.Model(&models.BootstrapTask{}).Where("id = ?", taskID).Updates(updates).Error
}

func (s *bootstrapService) updateRecord(recordID uint, updates map[string]interface{}) error {
	return s.db.Model(&models.BootstrapRecord{}).Where("id = ?", recordID).Updates(updates).Error
}

func (s *bootstrapService) logTaskError(taskID uint, message string, err error) {
	log.Printf("WARNING: bootstrap task %d: %s: %v", taskID, message, err)
}

func findAnsibleAssets() (string, string) {
	candidates := candidateRoots()
	for _, root := range candidates {
		rolePath := filepath.Join(root, "scripts", "ansible", "roles", "fluent_manager_agent")
		if info, err := os.Stat(rolePath); err == nil && info.IsDir() {
			binaryPath := filepath.Join(root, "scripts", "ansible", "files", "fluent-manager-agent")
			if info, err := os.Stat(binaryPath); err == nil && !info.IsDir() {
				return rolePath, binaryPath
			}
			return rolePath, ""
		}
	}
	return "", ""
}

func defaultAgentBinaryHint() string {
	_, binaryPath := findAnsibleAssets()
	if binaryPath != "" {
		return binaryPath
	}
	return "scripts/ansible/files/fluent-manager-agent"
}

func candidateRoots() []string {
	var roots []string
	if cwd, err := os.Getwd(); err == nil {
		roots = append(roots, cwd, filepath.Dir(cwd))
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		roots = append(roots, exeDir, filepath.Dir(exeDir), filepath.Dir(filepath.Dir(exeDir)))
	}
	unique := make([]string, 0, len(roots))
	seen := map[string]struct{}{}
	for _, root := range roots {
		if root == "" {
			continue
		}
		root = filepath.Clean(root)
		if _, ok := seen[root]; ok {
			continue
		}
		seen[root] = struct{}{}
		unique = append(unique, root)
	}
	return unique
}
