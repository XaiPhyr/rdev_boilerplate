package utils

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"gopkg.in/yaml.v2"
)

type (
	Config struct {
		Server   ServerConfig   `yaml:"server"`
		Database DatabaseConfig `yaml:"database"`
		Frontend FrontendConfig `yaml:"frontend"`
		Env      string         `yaml:"env"`
		SMTP     SMTPConfig     `yaml:"smtp"`
		Redis    RedisConfig    `yaml:"redis"`
	}

	ServerConfig struct {
		Host     string `yaml:"host"`
		Endpoint string `yaml:"endpoint"`
		JwtKey   string `yaml:"jwt_key"`
	}

	DatabaseConfig struct {
		DSN      string `yaml:"dsn"`
		BunDebug bool   `yaml:"bundebug"`
	}

	FrontendConfig struct {
		Source string `yaml:"src"`
	}

	SMTPConfig struct {
		Host string `yaml:"host"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
		Port int    `yaml:"port"`
	}

	RedisConfig struct {
		Address string `yaml:"address"`
	}
)

var (
	cfg = InitConfig()
	db  *bun.DB
)

func InitConfig() Config {
	var cfg Config

	filePath := "./conf/config.yml"

	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error: %s", err)
	}

	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)

	if err != nil {
		log.Printf("Error: %s", err)
	}

	return cfg
}

func InitDB() *bun.DB {
	if db != nil {
		return db
	}

	dbConf := cfg.Database

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dbConf.DSN)))
	sqldb.SetMaxOpenConns(25)
	sqldb.SetMaxIdleConns(5)
	sqldb.SetConnMaxLifetime(1 * time.Hour)
	sqldb.SetConnMaxIdleTime(30 * time.Minute)

	db = bun.NewDB(sqldb, pgdialect.New())

	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	log.Println("✅ Database connection established successfully.")

	if dbConf.BunDebug {
		log.Println("⚙️  Bun debug query hook is enabled.")
	} else {
		log.Println("⚙️  Bun debug query hook is disabled.")
	}

	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithEnabled(dbConf.BunDebug),
		bundebug.WithVerbose(dbConf.BunDebug),
	))

	return db
}
