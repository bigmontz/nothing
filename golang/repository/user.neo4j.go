package repository

import (
	"fmt"
	"github.com/bigmontz/nothing/ioutils"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"strconv"
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

const updatePasswordQuery = `
MATCH (user:User {password: $current})
WHERE ID(user) = $id
SET user.password = $new, user.updatedAt = $updatedAt
RETURN user{id: $id} AS user
`

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
		return extractUserFromResult(result)
	})
	if err != nil {
		return
	}
	return res.(*User), nil
}

func (u *userNeo4jRepository) FindById(rawUserId interface{}) (result *User, err error) {
	userId, err := strconv.Atoi(rawUserId.(string))
	if err != nil {
		return nil, userError{err: fmt.Errorf("invalid user ID: %w", err)}
	}
	session := u.driver.NewSession(neo4j.SessionConfig{})
	defer func() {
		err = ioutils.SafeClose(err, session)
	}()
	res, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, txFuncErr := tx.Run(findByIdQuery, map[string]interface{}{"id": userId})
		if txFuncErr != nil {
			return nil, txFuncErr
		}
		return extractUserFromResult(result)
	})
	if err != nil {
		return
	}
	return res.(*User), nil
}

func (u *userNeo4jRepository) UpdatePassword(rawUserId interface{}, request *PasswordUpdate) (*User, error) {
	userId, err := strconv.Atoi(rawUserId.(string))
	if err != nil {
		return nil, userError{err: fmt.Errorf("invalid user ID: %w", err)}
	}
	session := u.driver.NewSession(neo4j.SessionConfig{})
	defer func() {
		err = ioutils.SafeClose(err, session)
	}()
	res, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, txFuncErr := tx.Run(updatePasswordQuery, map[string]interface{}{
			"id":        userId,
			"current":   request.Current,
			"new":       request.New,
			"updatedAt": time.Now().In(time.UTC),
		})
		if txFuncErr != nil {
			return nil, txFuncErr
		}
		if !result.Next() {
			return nil, userError{
				err:      fmt.Errorf("could not find user"),
				notFound: true,
			}
		}
		user, found := result.Record().Get("user")
		if !found {
			return nil, userError{
				err:      fmt.Errorf("unexpected query result"),
				notFound: true,
			}
		}
		return &User{
			Id: (user.(map[string]interface{})["id"]).(int64),
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return res.(*User), nil
}

func (u *userNeo4jRepository) Close() error {
	return u.driver.Close()
}

func extractUserFromResult(result neo4j.Result) (interface{}, error) {
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
	creationTime := user.Props["createdAt"].(time.Time)
	updateTime := user.Props["updatedAt"].(time.Time)
	return &User{
		Username:  user.Props["username"].(string),
		Name:      user.Props["name"].(string),
		Age:       uint(user.Props["age"].(int64)),
		Surname:   user.Props["surname"].(string),
		Password:  user.Props["password"].(string),
		CreatedAt: &creationTime,
		UpdatedAt: &updateTime,
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
