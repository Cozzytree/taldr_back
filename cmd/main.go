package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/Cozzytree/taldrBack/internal/server"
	"github.com/joho/godotenv"
)

func gracefullShutdown(apiServer *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	stopDockerContainer("taldr")

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func stopDockerContainer(containerName string) {
	// Command to stop the Docker container
	cmd := exec.Command("sudo", "docker", "stop", containerName)

	// Run the command and check for errors
	if err := cmd.Run(); err != nil {
		log.Printf("Error stopping Docker container %s: %v", containerName, err)
	} else {
		log.Printf("Docker container %s stopped successfully", containerName)
	}
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}
	done := make(chan bool)

	server := server.NewServer()

	go gracefullShutdown(server, done)

	fmt.Printf("server started on address: %v \n", server.Addr)

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

	<-done
	fmt.Println("Gracefull shutdowned")
}
