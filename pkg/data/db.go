package data

import (
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kmgreen2/agglo/pkg/util"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Database is the interface to be implemented by all relational databases
type Database interface {
	Get() *gorm.DB
	Close() error
	Ping() error
}

// TestDatabase is the interface to be implemented by all databases that include mocking functionality
type TestDatabase interface {
	Database
	Mock() sqlmock.Sqlmock
}

// Database corresponds to a connection to a single database instance
type DatabaseImpl struct {
	gormDb *gorm.DB
	db *sql.DB
	mock sqlmock.Sqlmock
}

// NewDatabase will return a Database object associated with the provided config, or an error if something went wrong
func NewDatabaseImpl(config *DatabaseConfig) (*DatabaseImpl, error) {
	var err error
	if config.databaseType == MockDatabase {
		database := &DatabaseImpl{}
		database.db, database.mock, err = sqlmock.New()
		if err != nil {
			return nil, err
		}
		database.gormDb, err = gorm.Open(postgres.New(postgres.Config{
			Conn: database.db,
		}), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		return database, nil
	} else if config.databaseType == PostgresDatabase {
		database := &DatabaseImpl{}
		database.gormDb, err = gorm.Open(postgres.Open(config.dataSourceName), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		return database, nil
	}
	return nil, util.NewInvalidError(fmt.Sprintf("NewDatabase - invalid database type: %s",
		config.databaseType))
}

// Get will return the underlying database connection object
func (db *DatabaseImpl) Get() *gorm.DB {
	return db.gormDb
}

// Close will close the underlying database connection
func (db *DatabaseImpl) Close() error {
	sqlDB, err := db.gormDb.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping will test the underlying database connection
func (db *DatabaseImpl) Ping() error {
	return db.gormDb.Raw("SELECT 1").Error
}

// Mock will return the underlying mock expectation object for a mocked database
// This is used in testing to check expectations
func (db *DatabaseImpl) Mock() sqlmock.Sqlmock {
	return db.mock
}
