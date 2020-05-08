package conf

import (
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

type SrvConfig struct {
	Host            string
	Port            int
	DebugHost       string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

func NewSrvConfig(host string, port int) SrvConfig {
	return SrvConfig{
		Host:            host,
		Port:            port,
		ReadTimeout:     time.Second * 5,
		WriteTimeout:    time.Second * 5,
		ShutdownTimeout: time.Second * 5,
	}
}

func (sc SrvConfig) Addr() string {
	return sc.Host + ":" + strconv.Itoa(sc.Port)
}

type AuthConfig struct {
	KeyID          string
	PrivateKeyFile string
	Algorithm      string
}

type Config struct {
	Server SrvConfig
	Auth   AuthConfig
	Db     db.Config
	Args   Args
}

func NewConfig() *Config {
	initCliFlags()

	return &Config{
		Server: srvConfig(),
		Auth:   authConfig(),
		Db:     dbConfig(),
		Args:   NewConfigArgs(os.Args[1:]),
	}
}

func srvConfig() SrvConfig {
	return NewSrvConfig(
		viper.GetString("host"),
		viper.GetInt("port"))
}

func authConfig() AuthConfig {
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
