package config

import (
	"context"
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
		ioutils.ReadEnv("MONGODB_USER", "mongodb"),
		ioutils.ReadEnv("MONGODB_ADDRESS", "localhost"),
	)
	return mongo.Connect(context.Background(), options.Client().ApplyURI(url))
}
