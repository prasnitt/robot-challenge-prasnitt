package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/prasnitt/robot-challenge-prasnitt/internal/api"
)

func main() {
	fmt.Println("Robot Warehouse System Starting...")

	// Initialize the Gin router
	router := gin.Default()

	// Setup API routes
	api.SetupRouter(router)

	// Start the server
	if err := router.Run(":8080"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}

}
