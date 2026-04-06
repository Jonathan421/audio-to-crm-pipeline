package main

import (
	"log"
	"os"
	"voiceline-mvp/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, falling back to system environment variables")
	}

	// 2. Fail-fast check to ensure critical API keys are present before starting
	if os.Getenv("OPENAI_API_KEY") == "" || os.Getenv("HUBSPOT_TOKEN") == "" {
		log.Fatal("CRITICAL ERROR: OPENAI_API_KEY or HUBSPOT_TOKEN is missing in the environment")
	}

	// 3. Initialize the Gin router with default middleware (logger and recovery)
	router := gin.Default()

	// 4. Simple health-check endpoint (useful for debugging and deployment readiness)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong - Server is running!",
		})
	})

	// 5. Main endpoint for the audio-to-CRM pipeline
	router.POST("/api/v1/note", handlers.HandleAudioUpload)

	// 6. Define the port and start the HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server is starting on port %s...", port)
	router.Run(":" + port)
}
