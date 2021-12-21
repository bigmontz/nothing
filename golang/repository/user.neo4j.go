package repository

import (
	"fmt"
	"github.com/bigmontz/nothing/ioutils"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"time"
)

type userNeo4jRepository struct {
	driver neo4j.Driver
}

func NewUserNeo4jRepository(driver neo4j.Driver) UserRepository {
	return &userNeo4jRepository{driver: driver}
}

const createQuery = `
CREATE (user:User {
	username: $username,
	name: $name,
	surname: $surname,
	age: $age,
	password: $password,
	createdAt: $createdAt,
	updatedAt: $updatedAt })
RETURN user`

const findByIdQuery = `MATCH (user:User) WHERE ID(user) = $id RETURN user`

func (u *userNeo4jRepository) Create(user *User) (result *User, err error) {
	session := u.driver.NewSession(neo4j.SessionConfig{})
	defer func() {
		err = ioutils.SafeClose(err, session)
	}()
	res, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, txFuncErr := tx.Run(createQuery, userParameters(user))
		if txFuncErr != nil {
			return nil, txFuncErr
		}
		return extractUser(result)
	})
	if err != nil {
		return
	}
	return res.(*User), nil
}

func (u *userNeo4jRepository) FindById(userId int64) (result *User, err error) {
	session := u.driver.NewSession(neo4j.SessionConfig{})
	defer func() {
		err = ioutils.SafeClose(err, session)
	}()
	res, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, txFuncErr := tx.Run(findByIdQuery, map[string]interface{}{"id": userId})
		if txFuncErr != nil {
			return nil, txFuncErr
		}
		return extractUser(result)
	})
	if err != nil {
		return
	}
	return res.(*User), nil
}

func (u *userNeo4jRepository) Close() error {
	return u.driver.Close()
}

func extractUser(result neo4j.Result) (interface{}, error) {
	record, txFuncErr := result.Single()
	if txFuncErr != nil {
		return nil, txFuncErr
	}
	user, found := record.Get("user")
	if !found {
		return nil, fmt.Errorf("could not find user")
	}
	return userToMap(user.(neo4j.Node)), nil
}

func userToMap(user neo4j.Node) *User {
	return &User{
		Username:  user.Props["username"].(string),
		Name:      user.Props["name"].(string),
		Age:       uint(user.Props["age"].(int64)),
		Surname:   user.Props["surname"].(string),
		Password:  user.Props["password"].(string),
		CreatedAt: user.Props["createdAt"].(time.Time),
		UpdatedAt: user.Props["updatedAt"].(time.Time),
		Id:        user.Id,
	}
}

func userParameters(user *User) map[string]interface{} {
	now := time.Now().In(time.UTC)
	return map[string]interface{}{
		"username":  user.Username,
		"name":      user.Name,
		"surname":   user.Surname,
		"age":       user.Age,
		"password":  user.Password,
		"createdAt": now,
		"updatedAt": now,
	}
}
