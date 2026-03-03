package database

import (
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// In order to connect to Postgresql you need to fill out all the fields.
//
// To connect to sqlite, you just need to specify "sqlite" driver.
// By default it will use in-memory database. You can provide CLEARNODE_DATABASE_NAME to use the file.
type DatabaseConfig struct {
	URL      string `env:"CLEARNODE_DATABASE_URL" env-default:""`
	Name     string `env:"CLEARNODE_DATABASE_NAME" env-default:""`
	Schema   string `env:"CLEARNODE_DATABASE_SCHEMA" env-default:""`
	Driver   string `env:"CLEARNODE_DATABASE_DRIVER" env-default:"postgres"`
	Username string `env:"CLEARNODE_DATABASE_USERNAME"  env-default:"postgres"`
	Password string `env:"CLEARNODE_DATABASE_PASSWORD" env-default:"your-super-secret-and-long-postgres-password"`
	Host     string `env:"CLEARNODE_DATABASE_HOST" env-default:"localhost"`
	Port     string `env:"CLEARNODE_DATABASE_PORT" env-default:"5432"`
	Retries  int    `env:"CLEARNODE_DATABASE_RETRIES" env-default:"5"`

	// Connection pool settings
	MaxOpenConns    int `env:"CLEARNODE_DATABASE_MAX_OPEN_CONNS" env-default:"100"`
	MaxIdleConns    int `env:"CLEARNODE_DATABASE_MAX_IDLE_CONNS" env-default:"25"`
	ConnMaxLifetime int `env:"CLEARNODE_DATABASE_CONN_MAX_LIFETIME_SEC" env-default:"300"`
	ConnMaxIdleTime int `env:"CLEARNODE_DATABASE_CONN_MAX_IDLE_TIME_SEC" env-default:"60"`
}

func ConnectToDB(cnf DatabaseConfig, embedMigrations embed.FS) (*gorm.DB, error) {
	switch cnf.Driver {
	case "postgres":
		return connectToPostgresql(cnf, embedMigrations)
	case "sqlite", "":
		return connectToSqlite(cnf)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cnf.Driver)
	}
}

func connectToPostgresql(cnf DatabaseConfig, embedMigrations embed.FS) (*gorm.DB, error) {
	log.Println("connecting to Postgresql")
	// Create schema if not exists
	if err := ensurePostgresqlSchema(cnf); err != nil {
		return nil, fmt.Errorf("failed to ensure Postgresql schema: %w", err)
	}

	// Apply migrations
	if err := migratePostgres(cnf, embedMigrations); err != nil {
		return nil, fmt.Errorf("failed to apply Postgresql migrations: %w", err)
	}

	// Connect to db
	dsn, err := postgresqlDbUrl(cnf)
	if err != nil {
		return nil, err
	}
	dial := postgres.Open(dsn)

	db, err := gorm.Open(dial, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: cnf.Schema + ".", // schema name
		},
		// Don't prepare statements, as it can cause issues with some Postgresql proxies like pgbouncer in transaction pooling mode.
		PrepareStmt: false,
		// Reduce log noise in production.
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cnf.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cnf.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cnf.ConnMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cnf.ConnMaxIdleTime) * time.Second)

	log.Printf("DB pool configured: maxOpen=%d, maxIdle=%d, maxLifetime=%ds, maxIdleTime=%ds",
		cnf.MaxOpenConns, cnf.MaxIdleConns, cnf.ConnMaxLifetime, cnf.ConnMaxIdleTime)

	return db, nil
}

func connectToSqlite(cnf DatabaseConfig) (*gorm.DB, error) {
	var dsn string
	if cnf.Name != "" {
		log.Println("connecting to sqlite")
		dsn = fmt.Sprintf("file:%s?cache=shared", cnf.Name)
	} else {
		log.Println("connecting to in-memory sqlite")
		dsn = "file::memory:?cache=shared"
	}
	dial := sqlite.Open(dsn)

	db, err := gorm.Open(dial, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: cnf.Schema + ".", // schema name
		},
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, err
	}

	// Migrate sqlite
	migrateSqlite(db)

	log.Println("Successfully auto-migrated")

	return db, nil
}

func postgresqlDbUrl(cnf DatabaseConfig) (string, error) {
	switch cnf.Driver {
	case "postgres":
		dsn := fmt.Sprintf(
			"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
			cnf.Username, cnf.Password, cnf.Host, cnf.Port, cnf.Name,
		)

		if cnf.Schema != "" {
			dsn = fmt.Sprintf("%s search_path=%s", dsn, cnf.Schema)
		}

		return dsn, nil

	default:
		return "", fmt.Errorf("unsupported driver: %s", cnf.Driver)
	}
}

func ensurePostgresqlSchema(cnf DatabaseConfig) error {
	if cnf.Schema == "" {
		log.Println("No schema specified, skipping schema creation")
		return nil
	}

	log.Println("creating schema")
	dbConf := cnf
	dbConf.Schema = ""
	dsn, err := postgresqlDbUrl(dbConf)
	if err != nil {
		return err
	}

	db, err := sqlx.Connect(dbConf.Driver, dsn)
	if err != nil {
		return err
	}

	queryDbCheck := fmt.Sprintf("SELECT 1 FROM information_schema.schemata WHERE schema_name='%s'", cnf.Schema)
	if res, err := db.Exec(queryDbCheck); err != nil {
		return fmt.Errorf("error while checking schema existance: %s", err.Error())
	} else if rows, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("error while checking schema existance: %s", err.Error())
	} else if rows > 0 {
		log.Printf("Schema already exists: %s\n", cnf.Schema)
		return nil
	}

	if _, err = db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", cnf.Schema)); err != nil {
		return fmt.Errorf("error while creating schema: %s", err.Error())
	}

	log.Printf("Schema created: %s\n", cnf.Schema)
	return nil
}

func migratePostgres(cnf DatabaseConfig, embedMigrations embed.FS) error {
	dsn, err := postgresqlDbUrl(cnf)
	if err != nil {
		return err
	}

	db, err := goose.OpenDBWithDriver(cnf.Driver, dsn)
	if err != nil {
		return err
	}

	if cnf.Schema != "" {
		switch cnf.Driver {
		case "postgres":
			if _, err := db.Exec(fmt.Sprintf("SET search_path TO %s", cnf.Schema)); err != nil {
				return fmt.Errorf("failed to set search path: %v", err)
			}
		}
	}

	log.Println("Applying database migrations")
	goose.SetBaseFS(embedMigrations)
	if err := goose.Up(db, "config/migrations/"+cnf.Driver); err != nil {
		panic(err)
	}

	log.Println("Applied migrations")
	return nil
}

func migrateSqlite(db *gorm.DB) error {
	if err := db.AutoMigrate(&AppV1{}, &AppLedgerEntryV1{}, &Channel{}, &AppSessionV1{}, &ContractEvent{}, &BlockchainAction{}, &AppSessionKeyStateV1{}, &AppSessionKeyApplicationV1{}, &AppSessionKeyAppSessionIDV1{}, &UserBalance{}); err != nil {
		return err
	}
	return nil
}
