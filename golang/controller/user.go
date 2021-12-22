package controller

import (
	"encoding/json"
	"fmt"
	"github.com/bigmontz/nothing/repository"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type userController struct {
	userRepository repository.UserRepository
}

func NewUserController(userRepository repository.UserRepository) http.Handler {
	return &userController{
		userRepository: userRepository,
	}
}

func (u *userController) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if !strings.HasPrefix(request.RequestURI, "/user") {
		responseWriter.WriteHeader(http.StatusNotFound)
		return
	}
	switch request.Method {
	case "POST":
		u.create(responseWriter, request)
	case "GET":
		u.findById(responseWriter, request)
	default:
		responseWriter.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (u *userController) create(responseWriter http.ResponseWriter, request *http.Request) {
	user, err := unmarshalUser(request.Body)
	if err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Header().Add("Content-Type", "plain/text")
		_, _ = responseWriter.Write([]byte(err.Error()))
	}
	result, err := u.userRepository.Create(user)
	marshalUser(responseWriter, err, result)
}

func (u *userController) findById(responseWriter http.ResponseWriter, request *http.Request) {
	rawUserId := strings.TrimPrefix(request.RequestURI, "/user/")
	userId, err := strconv.ParseInt(rawUserId, 10, 64)
	if err != nil {
		responseWriter.WriteHeader(http.StatusBadRequest)
		responseWriter.Header().Add("Content-Type", "plain/text")
		_, _ = responseWriter.Write([]byte(fmt.Sprintf("invalid user ID: %s", err.Error())))
	}
	result, err := u.userRepository.FindById(userId)
	marshalUser(responseWriter, err, result)
}

func marshalUser(responseWriter http.ResponseWriter, err error, result *repository.User) {
	if err != nil {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		responseWriter.Header().Add("Content-Type", "plain/text")
		_, _ = responseWriter.Write([]byte(err.Error()))
		return
	}
	responseWriter.WriteHeader(http.StatusOK)
	responseWriter.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(responseWriter).Encode(result)
}

func unmarshalUser(body io.ReadCloser) (*repository.User, error) {
	var result repository.User
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
