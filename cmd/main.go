package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prasnitt/robot-challenge-prasnitt/internal/api"
	"github.com/prasnitt/robot-challenge-prasnitt/internal/robot"
)

func main() {
	log.Println("Robot Warehouse System Starting...")

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a buffered channel for robot tasks
	taskQueue := make(chan robot.RobotTask, 100) // Adjust buffer size as needed

	// Initialize the robot service
	robotService := robot.NewService(ctx, taskQueue)

	// Start the robot service in a separate goroutine
	go robotService.Start()

	// Initialize the Gin router
	router := gin.Default()

	// Setup API routes
	api.SetupRouter(router, robotService)

	// Start the server
	// TODO: Change the port to a configurable value
	port := ":8080"
	log.Printf("Starting server on %s...\n", port)
	if err := router.Run(port); err != nil {
		log.Printf("Failed to start server: %v\n", err)
		return
	}
}
