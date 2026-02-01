package sqldatabase

import (
	"fmt"
	"log/slog"
	neturl "net/url"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Storage implements multiple interfaces required by [service] packages.
// E.g. [service.UserRepository], [service.OrderRepository], [service.BalanceRepository].
// It also implements [workers.OrderRepository] required for [service.OrderProcessor].
// Internally, it wraps GORM database connection.
type Storage struct {
	db *gorm.DB
}

// NewStorage creates an instance of [Storage] from a
// provided database connection string (DSN).
func NewStorage(databaseURI string) (*Storage, error) {
	driver, err := getDatabaseDriver(databaseURI)
	if err != nil {
		return nil, err
	}

	return NewStorageFromDriver(driver)
}

// NewStorageFromDriver creates an instance of [Storage] from a
// preconfigured database driver instance.
func NewStorageFromDriver(driver gorm.Dialector) (*Storage, error) {
	conn, err := gorm.Open(driver, getDatabaseConfig())
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}

	storage := &Storage{
		db: conn,
	}

	return storage, nil
}

// WithLogLevel returns a new instance of storage with a new
// database session configured with the given log level.
func (s *Storage) WithLogLevel(level logger.LogLevel) *Storage {
	db := s.db.Session(&gorm.Session{ //nolint:exhaustruct
		Logger: s.db.Logger.LogMode(level),
	})

	return &Storage{
		db: db,
	}
}

// Debug returns a new instance of storage with debug-enabled
// database session.
func (s *Storage) Debug() *Storage {
	db := s.db.Debug()

	return &Storage{
		db: db,
	}
}

//nolint:ireturn
func getDatabaseDriver(databaseURI string) (gorm.Dialector, error) {
	url, err := neturl.Parse(databaseURI)
	if err != nil {
		return nil, fmt.Errorf("cannot parse database uri (%s): %w", databaseURI, err)
	}

	switch url.Scheme {
	case "postgres", "postgresql":
		return postgres.Open(url.String()), nil
	case "sqlite":
		url.Scheme = ""

		return sqlite.Open(url.String()), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDatabaseDriver, url.Scheme)
	}
}

func getDatabaseConfig() *gorm.Config {
	var config gorm.Config

	config.Logger = getDatabaseLogger()

	return &config
}

//nolint:ireturn
func getDatabaseLogger() logger.Interface {
	return logger.NewSlogLogger(slog.Default(), logger.Config{
		LogLevel:                  logger.Error,
		SlowThreshold:             time.Second,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
		Colorful:                  false,
	})
}
