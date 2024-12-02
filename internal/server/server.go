package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Cozzytree/taldrBack/internal/database"
)

const (
	SAVE_INTERVAL = 15
)

type Server struct {
	port int

	db     database.Service
	shapeS *ShapeStore
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	s := NewStore()

	NewServer := &Server{
		port:   port,
		db:     database.New(),
		shapeS: s,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	// Start the periodic task in a goroutine
	go func() {
		ticker := time.NewTicker(SAVE_INTERVAL * time.Second)
		defer ticker.Stop() // Ensure the ticker stops when done

		for {
			select {
			case <-ticker.C:
				// log.Println("Saving shapes to database...")
				// Call the storeinDb method to save shapes to the database
				s.storeinDb(NewServer.db)
			}
		}
	}()

	return server
}
