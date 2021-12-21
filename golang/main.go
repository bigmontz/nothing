package main

import (
	"context"
	"fmt"
	"github.com/bigmontz/nothing/config"
	"github.com/bigmontz/nothing/controller"
	"github.com/bigmontz/nothing/ioutils"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	handlers := http.NewServeMux()
	userRepository, err := config.GetUserRepository()
	ioutils.PanicOnError(err)
	userController := controller.NewUserController(userRepository)
	handlers.Handle(`/user`, userController)
	handlers.Handle(`/user/`, userController)

	listener, err := net.Listen("tcp", ":3001")
	ioutils.PanicOnError(err)
	server := http.Server{
		Addr:    listener.Addr().String(),
		Handler: handlers,
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		<-signals
		fmt.Println("Graceful shutdown started")
		_ = userRepository.Close()
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		_ = server.Shutdown(timeout)
		fmt.Println("... done!")
		done <- true
	}()
	if err = server.Serve(listener); err != http.ErrServerClosed {
		panic(err)
	}
	<-done
}
