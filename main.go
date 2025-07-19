package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prasnitt/robot-challenge-prasnitt/internal/api"
	"github.com/prasnitt/robot-challenge-prasnitt/internal/robot"

	_ "github.com/prasnitt/robot-challenge-prasnitt/docs" // Import the generated docs package
	swaggerFiles "github.com/swaggo/files"                // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger"            // gin-swagger middleware
)

// @title Robot Warehouse System
// @version 1.0
// @description This is a REST API for managing robot tasks in a warehouse system.
// @contact.name Prashant Agrawal
// @contact.email prashant.eee.nitt@gmail.com

// @host localhost:8080
// @BasePath /api/v1
func main() {
	log.Println("Robot Warehouse System Starting...")

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a buffered channel for robot tasks
	maxNumTasks := 100 // Maximum number of tasks in the queue
	taskIdQueue := make(chan string, maxNumTasks)

	// Initialize the robot service
	robotService := robot.NewService(ctx, taskIdQueue)

	// Start the robot service in a separate goroutine
	go robotService.Start()

	// Initialize the Gin router
	router := gin.Default()

	// Setup API routes
	api.SetupRouter(router, robotService)

	// Swagger documentation route
	// The url points to the API definition (docs.json or docs.yaml)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start the server
	// TODO: Change the port to a configurable value
	port := ":8080"
	log.Printf("Starting server on %s...\n", port)
	if err := router.Run(port); err != nil {
		log.Printf("Failed to start server: %v\n", err)
		return
	}
}
