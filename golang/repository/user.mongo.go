package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"time"
)

type userMongoRepository struct {
	client *mongo.Client
}

func NewUserMongoRepository(driver *mongo.Client) UserRepository {
	return &userMongoRepository{client: driver}
}

func (u *userMongoRepository) Create(user *User) (*User, error) {
	result, err := u.userCollection().InsertOne(context.Background(), userToDocument(user))
	if err != nil {
		return nil, err
	}
	return u.FindById(result.InsertedID)
}

func (u *userMongoRepository) FindById(userId interface{}) (*User, error) {
	objectId, err := asObjectId(userId)
	if err != nil {
		return nil, err
	}
	byObjectId := bson.M{"_id": objectId}
	result := u.userCollection().FindOne(context.Background(), byObjectId)
	var document bson.M
	if err = result.Decode(&document); err != nil {
		return nil, err
	}
	return documentToUser(document), nil
}

func (u *userMongoRepository) UpdatePassword(userId interface{}, passwordUpdate *PasswordUpdate) (*User, error) {
	panic("implement me")
}

func (u *userMongoRepository) Close() error {
	return u.client.Disconnect(context.Background())
}

func (u *userMongoRepository) userCollection() *mongo.Collection {
	return u.client.Database("admin").Collection("users")
}

func userToDocument(user *User) bson.M {
	now := time.Now()
	return map[string]interface{}{
		"username":  user.Username,
		"name":      user.Name,
		"age":       int64(user.Age),
		"surname":   user.Surname,
		"password":  user.Password,
		"createdAt": now,
		"updatedAt": now,
	}
}

func documentToUser(doc bson.M) *User {
	creationTime := doc["createdAt"].(primitive.DateTime).Time()
	updateTime := doc["updatedAt"].(primitive.DateTime).Time()
	return &User{
		Username:  doc["username"].(string),
		Name:      doc["name"].(string),
		Age:       uint(doc["age"].(int64)),
		Surname:   doc["surname"].(string),
		Password:  doc["password"].(string),
		CreatedAt: &creationTime,
		UpdatedAt: &updateTime,
		Id:        doc["_id"].(primitive.ObjectID).Hex(),
	}
}

func asObjectId(userId interface{}) (primitive.ObjectID, error) {
	if objectId, ok := userId.(primitive.ObjectID); ok {
		return objectId, nil
	}
	if hex, ok := userId.(string); ok {
		return primitive.ObjectIDFromHex(hex)
	}
	return primitive.NilObjectID,
		fmt.Errorf("unsupported user id type: %v", reflect.TypeOf(userId))
}
