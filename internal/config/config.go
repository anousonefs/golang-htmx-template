package config

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Masterminds/squirrel"
)

type Config struct {
	dbDriver    string
	dbHost      string
	dbPort      string
	dbUser      string
	dbPassword  string
	dbName      string
	assetDir    string
	baseUrl     string
	sessionName string

	appPort      string
	pasetoSecret []byte

	oneSignalApiKey string
	oneSignalAppID  string

	facebookAppID     string
	facebookAppSecret string

	discordClientID     string
	discordClientSecret string
}

func (c Config) FacebookAppID() string {
	return c.facebookAppID
}

func (c Config) FacebookAppSecret() string {
	return c.facebookAppSecret
}

func (c Config) DiscordClientID() string {
	return c.discordClientID
}

func (c Config) DiscordClientSecret() string {
	return c.discordClientSecret
}

func (c Config) PasetoSecret() []byte {
	return c.pasetoSecret
}

func (c Config) AppPort() string {
	return c.appPort
}

func (c Config) DBDriver() string {
	return c.dbDriver
}

func (c Config) AssetDir() string {
	return c.assetDir
}

func (c Config) BaseUrl() string {
	return c.baseUrl
}

func (c Config) SessionName() string {
	return c.sessionName
}

func (c Config) DSNInfo() string {
	timeoutOption := fmt.Sprintf("-c statement_timeout=%d", 10*time.Minute/time.Millisecond)
	return fmt.Sprintf("user='%s' password='%s' host='%s' port=%s dbname='%s' sslmode=disable options='%s'", c.dbUser, c.dbPassword, c.dbHost, c.dbPort, c.dbName, timeoutOption)
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func NewConfig() (config Config, err error) {
	config.dbDriver = GetEnv("DB_DRIVER", "postgres")
	config.dbHost = GetEnv("PGHOST", "127.0.0.1")
	config.dbPort = GetEnv("PGPORT", "5432")
	config.dbUser = os.Getenv("PGUSER")
	config.dbPassword = os.Getenv("PGSECRET")
	config.dbName = os.Getenv("PGDATABASE")
	config.baseUrl = os.Getenv("BASE_URL")
	config.sessionName = os.Getenv("SESSION_NAME")

	if config.baseUrl == "" {
		return config, errors.New("BASE_URL is empty")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, err
	}
	config.assetDir = GetEnv("ASSET_DIR", homeDir)
	config.appPort = GetEnv("PORT", "8080")

	config.pasetoSecret, err = hex.DecodeString(os.Getenv("PASETO_SECRET"))
	if len(config.pasetoSecret) != 32 {
		return config, err
	}

	config.oneSignalApiKey = os.Getenv("ONESIGNAL_REST_API_KEY")
	config.oneSignalAppID = os.Getenv("ONESIGNAL_APP_ID_KEY")

	config.facebookAppID = os.Getenv("FACEBOOK_APP_ID")
	if config.facebookAppID == "" {
		return config, errors.New("facebook app idis empty")
	}
	config.facebookAppSecret = os.Getenv("FACEBOOK_APP_SECRET")
	if config.facebookAppSecret == "" {
		return config, errors.New("facebook app secretis empty")
	}
	config.discordClientID = os.Getenv("DISCORD_CLIENT_ID")
	if config.discordClientID == "" {
		return config, errors.New("discord client id is empty")
	}
	config.discordClientSecret = os.Getenv("DISCORD_CLIENT_SECRET")
	if config.discordClientSecret == "" {
		return config, errors.New("discord client secret is empty")
	}

	return
}

func Psql() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}

func Mysql() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
}
