package models

import "time"

// BootstrapHost stores a reusable SSH target for Ansible-based bootstrap.
type BootstrapHost struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	Hostname                string    `gorm:"size:256;not null" json:"hostname"`
	IPAddress               string    `gorm:"size:64" json:"ip_address"`
	SSHPort                 int       `json:"ssh_port"`
	SSHUser                 string    `gorm:"size:128;not null" json:"ssh_user"`
	AuthType                string    `gorm:"size:32;not null" json:"auth_type"`
	PasswordEncrypted       string    `gorm:"type:text" json:"-"`
	PrivateKeyEncrypted     string    `gorm:"type:text" json:"-"`
	BecomePasswordEncrypted string    `gorm:"type:text" json:"-"`
	HasPassword             bool      `json:"has_password"`
	HasPrivateKey           bool      `json:"has_private_key"`
	HasBecomePassword       bool      `json:"has_become_password"`
	NodeUID                 string    `gorm:"size:128" json:"node_uid"`
	Labels                  string    `gorm:"type:text" json:"labels"`
	Description             string    `gorm:"size:512" json:"description"`
	ClusterID               *uint     `gorm:"index" json:"cluster_id"`
	Cluster                 *Cluster  `json:"cluster,omitempty"`
	CreatedBy               uint      `json:"created_by"`
	Creator                 *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// BootstrapTask tracks a host bootstrap/install task executed via Ansible.
type BootstrapTask struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Name              string     `gorm:"size:128;not null" json:"name"`
	Status            string     `gorm:"size:32;default:pending" json:"status"`
	Message           string     `gorm:"type:text" json:"message"`
	ClusterID         *uint      `gorm:"index" json:"cluster_id"`
	Cluster           *Cluster   `json:"cluster,omitempty"`
	FluentType        string     `gorm:"size:32;not null" json:"fluent_type"`
	InstallRuntime    bool       `json:"install_runtime"`
	ServerURL         string     `gorm:"size:512;not null" json:"server_url"`
	AgentBinarySource string     `gorm:"size:32;not null" json:"agent_binary_source"`
	AgentBinaryPath   string     `gorm:"size:1024" json:"agent_binary_path"`
	AgentDownloadURL  string     `gorm:"size:1024" json:"agent_download_url"`
	TotalHosts        int        `json:"total_hosts"`
	SuccessCount      int        `json:"success_count"`
	FailCount         int        `json:"fail_count"`
	StartedAt         *time.Time `json:"started_at"`
	FinishedAt        *time.Time `json:"finished_at"`
	CreatedBy         uint       `json:"created_by"`
	Creator           *User      `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// BootstrapRecord tracks the per-host result of a bootstrap task.
type BootstrapRecord struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	BootstrapTaskID uint           `gorm:"index" json:"bootstrap_task_id"`
	BootstrapHostID *uint          `gorm:"index" json:"bootstrap_host_id"`
	BootstrapHost   *BootstrapHost `gorm:"foreignKey:BootstrapHostID" json:"bootstrap_host,omitempty"`
	Hostname        string         `gorm:"size:256;not null" json:"hostname"`
	IPAddress       string         `gorm:"size:64" json:"ip_address"`
	SSHPort         int            `json:"ssh_port"`
	SSHUser         string         `gorm:"size:128;not null" json:"ssh_user"`
	AuthType        string         `gorm:"size:32;not null" json:"auth_type"`
	NodeUID         string         `gorm:"size:128" json:"node_uid"`
	Labels          string         `gorm:"type:text" json:"labels"`
	ClusterID       *uint          `gorm:"index" json:"cluster_id"`
	Cluster         *Cluster       `json:"cluster,omitempty"`
	Alias           string         `gorm:"size:64;index" json:"alias"`
	NodeID          *uint          `gorm:"index" json:"node_id"`
	Node            *Node          `gorm:"foreignKey:NodeID" json:"node,omitempty"`
	Status          string         `gorm:"size:32;default:pending" json:"status"`
	Message         string         `gorm:"type:text" json:"message"`
	OutputExcerpt   string         `gorm:"type:text" json:"output_excerpt"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}
