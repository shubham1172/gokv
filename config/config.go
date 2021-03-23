package config

import (
	"github.com/spf13/viper"
	"strings"
)

// Configuration for gokv
type Configuration struct {
	Server   ServerConfiguration
	Logging  LoggingConfiguration
	Database DatabaseConfiguration
}

type ServerConfiguration struct {
	Address string
}

type LoggingConfiguration struct {
	LogType     string
	LogFileName string
}

type DatabaseConfiguration struct {
	DBName    string
	Host      string
	User      string
	Password  string
	SslStatus string
}

// GetConfiguration loads the app configuration from a given configFileName
func GetConfiguration() (*Configuration, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("gokv")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	viper.SetDefault("server.address", ":8000")
	viper.SetDefault("logging.logtype", "file")
	viper.SetDefault("logging.logfilename", "transactions.log")
	viper.SetDefault("database.dbname", "postgres")
	viper.SetDefault("database.host", "postgres")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.sslstatus", "disable")

	config := &Configuration{}
	err = viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
