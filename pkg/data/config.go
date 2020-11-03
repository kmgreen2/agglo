package data

import "github.com/kmgreen2/agglo/pkg/config"

type DatabaseType string

const (
	MockDatabase DatabaseType = "sqlmock"
	PostgresDatabase DatabaseType = "postgres"
)

// DatabaseConfig an in-memory object that represents the configuration for all object stores
type DatabaseConfig struct {
	*config.ConfigBase
	databaseType DatabaseType
	dataSourceName string
}

// NewDatabaseConfig will reconcile configuration values from the environment and return a complete config object
func NewDatabaseConfig(baseConfig *config.ConfigBase) (*DatabaseConfig, error) {
	var databaseType string
	dbConfig := &DatabaseConfig{
		ConfigBase: baseConfig,
	}
	v, err := dbConfig.GetAndValidate("databaseType", config.NotNil)
	if err != nil {
		return nil, err
	}

	err = config.GetStringOrError(v, &databaseType)
	if err != nil {
		return nil, err
	}
	dbConfig.databaseType = DatabaseType(databaseType)

	v, err = dbConfig.GetAndValidate("dataSourceName", config.NotNil)
	if err != nil {
		return nil, err
	}

	err = config.GetStringOrError(v, &dbConfig.dataSourceName)
	if err != nil {
		return nil, err
	}

	return dbConfig, nil
}

// SetDatabaseType setter for databaseType
func (dbConfig *DatabaseConfig) SetDatabaseType(databaseType DatabaseType) error {
	dbConfig.databaseType = databaseType
	return nil
}

// GetDatabaseType getter for databaseType
func (dbConfig *DatabaseConfig) GetDatabaseType() DatabaseType {
	return dbConfig.databaseType
}

// SetDSN setter for dataSourceName
func (dbConfig *DatabaseConfig) SetDSN(dsn string) error {
	dbConfig.dataSourceName = dsn
	return nil
}

// GetDSN getter for dataSourceName
func (dbConfig *DatabaseConfig) GetDSN() string {
	return dbConfig.dataSourceName
}

