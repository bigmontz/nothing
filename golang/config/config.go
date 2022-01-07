package config

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/bigmontz/nothing/ioutils"
	"github.com/bigmontz/nothing/repository"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetUserRepository() (repository.UserRepository, error) {
	dbType := ioutils.ReadEnv("DB_TYPE", "neo4j")
	switch dbType {
	case "neo4j":
		driver, err := neo4jDriver()
		if err != nil {
			return nil, err
		}
		return repository.NewUserNeo4jRepository(driver), nil
	case "postgres":
		driver, err := postgresDriver()
		if err != nil {
			return nil, err
		}
		return repository.NewUserPostgresRepository(driver), nil
	case "mongodb":
		driver, err := mongoDriver()
		if err != nil {
			return nil, err
		}
		return repository.NewUserMongoRepository(driver), nil
	case "cockroachdb":
		driver, err := cockroachDriver()
		if err != nil {
			return nil, err
		}
		if err := initCockroachTable(driver); err != nil {
			return nil, err
		}
		return repository.NewUserCockroachRepository(driver), nil
	default:
		return nil, fmt.Errorf("unsupported DB type %s", dbType)
	}
}

func neo4jDriver() (neo4j.Driver, error) {
	return neo4j.NewDriver(
		ioutils.ReadEnv("NEO4J_URL", "neo4j://localhost"),
		neo4j.BasicAuth(
			ioutils.ReadEnv("NEO4J_USER", "neo4j"),
			ioutils.ReadEnv("NEO4J_PASSWORD", "pass"),
			"",
		),
	)
}

func postgresDriver() (*pgxpool.Pool, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s",
		ioutils.ReadEnv("POSTGRES_USER", "postgres"),
		ioutils.ReadEnv("POSTGRES_PASSWORD", "postgres"),
		ioutils.ReadEnv("POSTGRES_URL", "localhost"),
	)
	return pgxpool.Connect(context.Background(), url)
}

func mongoDriver() (*mongo.Client, error) {
	url := fmt.Sprintf("mongodb://%s:%s@%s",
		ioutils.ReadEnv("MONGODB_USER", "mongodb"),
		ioutils.ReadEnv("MONGODB_PASSWORD", "mongodb"),
		ioutils.ReadEnv("MONGODB_ADDRESS", "localhost"),
	)
	return mongo.Connect(context.Background(), options.Client().ApplyURI(url))
}

func cockroachDriver() (*sql.DB, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		ioutils.ReadEnv("COCKROACH_USER", "admin"),
		ioutils.ReadEnv("COCKROACH_USER", "cockroach"),
		ioutils.ReadEnv("COCKROACH_URL", "localhost"),
		ioutils.ReadEnv("COCKROACH_PORT", "26257"),
		ioutils.ReadEnv("COCKROACH_DATABASE", "postgres"),
	)
	return sql.Open("postgres", url)
}

func initCockroachTable(driver *sql.DB) error {
	_, err := driver.Exec("CREATE TABLE IF NOT EXISTS users (" +
		"id SERIAL PRIMARY KEY," +
		"username VARCHAR(255) NOT NULL," +
		"name VARCHAR(255) NOT NULL," +
		"surname VARCHAR(255) NOT NULL," +
		"password VARCHAR(255) NOT NULL," +
		"age INTEGER NOT NULL," +
		"created_at TIMESTAMP NOT NULL DEFAULT NOW(), " +
		"updated_at TIMESTAMP NOT NULL DEFAULT NOW()" +
		");")
	return err
}
