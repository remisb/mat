package conf

import (
	"github.com/remisb/mat/internal/conf"
	"github.com/remisb/mat/internal/db"
	"github.com/remisb/mat/internal/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strings"
	"sync"
)

const (
	configFileName = "config"
	envPrefix      = "mat"
)

var initConfigOnce sync.Once

// Config struct is used to store current application configuration.
type Config struct {
	Db   db.Config
	Args conf.Args
}

// NewConfig initializes and returns newly created Config struct.
func NewConfig() *Config {
	initCliFlags()

	return &Config{
		Db:   dbConfig(),
		Args: conf.NewConfigArgs(os.Args[1:]),
	}
}

func dbConfig() db.Config {
	return db.Config{
		Name:       viper.GetString("db-name"),
		Host:       viper.GetString("db-host"),
		Port:       viper.GetString("db-port"),
		User:       viper.GetString("db-user"),
		Password:   viper.GetString("db-password"),
		DisableTLS: viper.GetBool("db-tls-off"),
	}
}

func initCliFlags() {
	initConfigOnce.Do(func() {
		// setup cli flags
		pflag.CommandLine.IntP("port", "p", 8080, "api service port")

		// db config flags
		pflag.String("db-name", "postgres", "Database name")
		pflag.String("db-host", "localhost", "Database host")
		pflag.String("db-port", "5432", "Database port")
		pflag.String("db-user", "postgres", "Database user")
		pflag.String("db-password", "postgres", "Database password")
		pflag.Bool("db-tls-off", true, "Database disable TLS")
		pflag.Parse()

		if err := viper.BindPFlags(pflag.CommandLine); err != nil {
			log.Sugar.Errorf("bind CLI flags error: %v", err)
		}

		// setup environment variables
		viper.SetEnvPrefix(envPrefix)
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

		bindEnv("port")
		bindEnv("db-host")
		bindEnv("db-port")
		bindEnv("db-name")
		bindEnv("db-user")
		bindEnv("db-password")
		bindEnv("db-tls-off")

		// setup config file variables
		viper.SetConfigName(configFileName)
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")

		if err := viper.ReadInConfig(); err != nil {
			log.Sugar.Fatalf("Fatal error reading config file, err: %v", err)
		}
	})
}

func bindEnv(name string) {
	if err := viper.BindEnv(name); err != nil {
		log.Sugar.Errorf("bind env var %s error: %v", name, err)
	}
}
