package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/fluent-manager/fluent-manager/internal/config"
	"github.com/fluent-manager/fluent-manager/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SetupRequest holds the data needed to initialize the system.
type SetupRequest struct {
	// Database
	DBDriver   string `json:"db_driver"`
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`
	DBPath     string `json:"db_path"` // SQLite file path

	// Admin
	Username    string `json:"username"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

// TestDBRequest holds the fields for testing a database connection.
type TestDBRequest struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	Path     string `json:"path"` // SQLite
}

// SetupService handles first-run system initialization.
type SetupService interface {
	IsInitialized() (bool, error)
	Initialize(req SetupRequest) (*models.User, error)
	TestDBConnection(req TestDBRequest) error
	InitializeTarget(req SetupRequest, cfgPath string) (*models.User, error)
}

type setupService struct {
	db *gorm.DB
}

// NewSetupService creates a new SetupService.
func NewSetupService(db *gorm.DB) SetupService {
	return &setupService{db: db}
}

func (s *setupService) IsInitialized() (bool, error) {
	var count int64
	if err := s.db.Model(&models.User{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// BuildDSN constructs a DSN string from individual connection parameters.
func BuildDSN(driver, host string, port int, user, password, dbName, path string) string {
	switch driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			user, password, host, port, dbName)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbName)
	default:
		if path == "" {
			return "fluent_manager.db"
		}
		return path
	}
}

func (s *setupService) TestDBConnection(req TestDBRequest) error {
	dsn := BuildDSN(req.Driver, req.Host, req.Port, req.User, req.Password, req.DBName, req.Path)
	db, err := models.OpenDB(req.Driver, dsn)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return sqlDB.Ping()
}

// InitializeTarget connects to the target database, runs migrations, seeds defaults,
// creates the admin user, saves config.yaml, and returns the created user.
func (s *setupService) InitializeTarget(req SetupRequest, cfgPath string) (*models.User, error) {
	if req.Username == "" {
		return nil, ErrInvalidArgument
	}
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	driver := req.DBDriver
	if driver == "" {
		driver = "sqlite"
	}
	dsn := BuildDSN(driver, req.DBHost, req.DBPort, req.DBUser, req.DBPassword, req.DBName, req.DBPath)

	// Initialize target database (migrate + seed)
	targetDB, err := models.InitDBWithConn(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize target database: %w", err)
	}

	// Create admin user on target DB
	user, err := createAdminUser(targetDB, req)
	if err != nil {
		return nil, err
	}

	// Save config.yaml with the chosen database settings
	cfg, err := config.Load(cfgPath)
	if err != nil {
		cfg = &config.Config{}
	}
	cfg.Database.Driver = driver
	cfg.Database.DSN = dsn
	if err := config.Save(cfgPath, cfg); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	// Close target DB connection (server will reconnect after restart)
	if sqlDB, err := targetDB.DB(); err == nil {
		sqlDB.Close()
	}

	return user, nil
}

// Initialize creates the admin user on the current (boot) database.
// Used when the user keeps the default SQLite without changing DB settings.
func (s *setupService) Initialize(req SetupRequest) (*models.User, error) {
	initialized, err := s.IsInitialized()
	if err != nil {
		return nil, err
	}
	if initialized {
		return nil, errors.New("system is already initialized")
	}

	if req.Username == "" {
		return nil, ErrInvalidArgument
	}
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	return createAdminUser(s.db, req)
}

func createAdminUser(db *gorm.DB, req SetupRequest) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		DisplayName:  req.DisplayName,
		PasswordHash: string(hash),
		AuthSource:   "local",
		IsActive:     true,
		LastLoginAt:  &now,
	}

	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}

	var adminRole models.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		return nil, errors.New("admin role not found; ensure database is properly migrated")
	}
	if err := db.Model(&user).Association("Roles").Append(&adminRole); err != nil {
		return nil, err
	}

	db.Preload("Roles.Permissions").First(&user, user.ID)
	return &user, nil
}
