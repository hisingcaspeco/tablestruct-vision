package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hising/tablemap/client"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: No .env file found")
	}

	// Get API Key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY is not set")
		os.Exit(1)
	}

	// Initialize OpenAI client
	openAIClient := client.NewOpenAIClient(apiKey)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Your React frontend
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/", homeView)
	r.POST("/upload", func(c *gin.Context) {
		uploadView(c, openAIClient)
	})

	err = r.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func homeView(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello, World!",
	})
}

func uploadView(c *gin.Context, openAIClient *client.OpenAIClient) {
	// Get the uploaded file
	// Get uploaded file
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Read file into buffer
	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Convert to Base64
	base64Image := base64.StdEncoding.EncodeToString(buffer.Bytes())

	// Send image to OpenAI
	responseText, err := openAIClient.SendImageToOpenAI(base64Image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image with OpenAI", "details": err.Error()})
		return
	}

	// Respond with OpenAI's analysis
	c.JSON(http.StatusOK, gin.H{
		"message":  "Image processed successfully",
		"response": responseText,
	})
}
