package conf

import (
	"github.com/remisb/mat/cmd/rest-api/internal/web"
	"github.com/remisb/mat/internal/conf"
	"github.com/remisb/mat/internal/db"
	"github.com/remisb/mat/internal/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	configFileName = "config"
	envPrefix      = "mat"
)

var initConfigOnce sync.Once

// SrvConfig has server configuration settings.
type SrvConfig struct {
	Host            string
	Port            int
	Log             string
	DebugHost       string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// NewSrvConfig factory function creates and initialize new SrvConfig.
func NewSrvConfig(host string, port int, log string) SrvConfig {
	return SrvConfig{
		Host:            host,
		Port:            port,
		Log:             log,
		ReadTimeout:     time.Second * 5,
		WriteTimeout:    time.Second * 5,
		ShutdownTimeout: time.Second * 5,
	}
}

// Addr returns server address in the form of Host:Port localhost:8080.
func (sc SrvConfig) Addr() string {
	return sc.Host + ":" + strconv.Itoa(sc.Port)
}

// AuthConfig structure stores authentication configuration settings.
type AuthConfig struct {
	KeyID          string
	PrivateKeyFile string
	Algorithm      string
}

// Config structure to store application configuration settings.
type Config struct {
	Server SrvConfig
	Auth   AuthConfig
	Db     db.Config
	Args   conf.Args
}

// NewConfig is a factory class initializes and creates new Config structure.
func NewConfig() *Config {
	initCliFlags()

	return &Config{
		Server: srvConfig(),
		Auth:   authConfig(),
		Db:     dbConfig(),
		Args:   conf.NewConfigArgs(os.Args[1:]),
	}
}

func srvConfig() SrvConfig {
	return NewSrvConfig(
		viper.GetString("host"),
		viper.GetInt("port"),
		viper.GetString("log"),
	)
}

func authConfig() AuthConfig {
	web.InitAuth()
	return AuthConfig{
		KeyID:          viper.GetString("auth-keyid"),
		PrivateKeyFile: viper.GetString("auth-keyfile"),
		Algorithm:      viper.GetString("auth-algorithm"),
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
		pflag.CommandLine.StringP("host", "h", "localhost", "api service host")
		pflag.CommandLine.IntP("port", "p", 8080, "api service port")
		pflag.CommandLine.StringP("log", "l", "restaurant-api.log", "api service log file")

		// auth config flags
		pflag.String("auth-keyid", "1", "Authenticator Key ID")
		pflag.String("auth-keyfile", "private.pem", "Authenticator private key file")
		pflag.String("auth-algorithm", "RS256", "Authenticator algorithm")

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

		// bind server conf
		bindEnv("port")
		bindEnv("log")

		// bind auth conf
		bindEnv("auth-keyid")
		bindEnv("auth-keyfile")
		bindEnv("auth-algorithm")

		// bind db conf
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
