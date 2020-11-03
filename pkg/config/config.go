package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os/exec"
	"strconv"
	"strings"
)

// ConfigBase contains the basic elements needed to fetch and validate configs
type ConfigBase struct {
	viperConfig *viper.Viper
}

// NoValidate is the a no-op validation function
func NoValidate(value interface{}) error {
	return nil
}

// NotNil will ensure the argument is not nil
func NotNil(value interface{}) error {
	if value == nil {
		return fmt.Errorf("Nil config value")
	}
	return nil
}

// GetStringOrError will load a string from value into dest or return
// an error if there is an issue reading the value
func GetStringOrError(value interface{}, dest *string) error {
	switch value.(type) {
	case string:
		*dest = value.(string)
		return nil
	}
	return NewConfigError("Value is not a string")
}

// GetIntOrError will load an int from value into dest or return
// an error if there is an issue reading the value
func GetIntOrError(value interface{}, dest *int) error {
	switch value.(type) {
	case int:
		*dest = value.(int)
		return nil
	case string:
		intValue, err := strconv.Atoi(value.(string))
		if err == nil {
			*dest = intValue
			return nil
		}
	}
	return NewConfigError("Value is not a valid int")
}

// GetProjectRoot will detect the root directory for getting env when doing local development
// ToDo(KMG): Explicitly specify path(s) to any config files.
func GetProjectRoot() (string, error) {
	pathToGit, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	if err != nil {
		return "", err
	}
	projectRoot, err := exec.Command("dirname", strings.TrimSpace(string(pathToGit))).Output()
	if err != nil {
		return "", err
	}
	projectRootStr := strings.Trim(string(projectRoot), " \n")
	return projectRootStr, err
}

// GetAndValidate will read a config value, validate using validate function and set the appropriate entry using setFunc
func (config *ConfigBase) GetAndValidate(key string,
	validate func(interface{}) error) (interface{}, error) {

	configValue := config.viperConfig.Get(key)

	err := validate(configValue)
	if err != nil {
		return nil, NewConfigError(
			fmt.Sprintf("Validation failed for key=%s!  Msg: %s", key, err.Error()))
	}

	return configValue, nil
}

func NewConfigBase() (*ConfigBase, error) {
	v := viper.New()

	baseConfig := &ConfigBase{
		viperConfig: v,
	}

	projectRoot, err := GetProjectRoot()
	if err == nil {
		v.AddConfigPath(projectRoot)
		v.SetConfigName(".env")
		err = v.ReadInConfig()
		if err != nil {
			return nil, err
		}
	}
	v.AutomaticEnv()
	return baseConfig, nil
}
