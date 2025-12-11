package utils

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"gopkg.in/yaml.v2"
)

type (
	Utility struct{}

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

func GetPermissions(fn func(*bun.SelectQuery) *bun.SelectQuery, uuid string, ctx *gin.Context, dest ...any) error {
	q := db.NewSelect().
		With("UserPermissions", userPermissionsUnion()).
		TableExpr("users AS u").
		ColumnExpr("JSON_ARRAYAGG(p.name) AS permissions").
		Join(`LEFT JOIN "UserPermissions" up ON up.user_id = u.id`).
		Join("LEFT JOIN permissions p ON p.id = up.permission_id").
		WhereGroup("AND", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Where("u.uuid = ?", uuid).
				Where("u.status = 'O'").
				Where("u.deleted_at IS NULL")
		})

	if fn != nil {
		q = fn(q)
	}

	return q.Scan(ctx, &dest)
}

func userPermissionsUnion() *bun.SelectQuery {
	rolePermsQuery := db.NewSelect().TableExpr("users AS u").
		ColumnExpr("u.id AS user_id, rp.permission_id").
		Join("LEFT JOIN user_roles ur ON ur.user_id = u.id AND ur.deleted_at IS NULL AND ur.status = 'O'").
		Join("LEFT JOIN role_permissions rp ON rp.role_id = ur.role_id AND rp.deleted_at IS NULL AND rp.status = 'O'")

	groupPermsQuery := db.NewSelect().TableExpr("users AS U").
		ColumnExpr("u.id AS user_id, gp.permission_id").
		Join("LEFT JOIN user_groups ug ON ug.user_id = u.id AND ug.deleted_at IS NULL AND ug.status = 'O'").
		Join("LEFT JOIN group_permissions gp ON gp.group_id = ug.group_id AND gp.deleted_at IS NULL AND gp.status = 'O'")

	return rolePermsQuery.Union(groupPermsQuery)
}
