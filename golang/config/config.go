package config

import (
	"fmt"
	"github.com/bigmontz/nothing/ioutils"
	"github.com/bigmontz/nothing/repository"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
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