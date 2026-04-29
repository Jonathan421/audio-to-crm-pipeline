package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"voiceline-mvp/services"

	"github.com/gin-gonic/gin"
)

// HandleAudioUpload orchestrates the entire AI-to-CRM pipeline:
// 1. Audio Upload -> 2. Whisper Transcription -> 3. GPT-4o Extraction -> 4. HubSpot CRM Sync
func HandleAudioUpload(c *gin.Context) {
	// 1. Extract the audio file from the multipart form request
	file, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No audio file found in the 'audio' form field"})
		return
	}

	// 2. Ensure a local temporary directory exists for processing the file
	tmpFolder := "./tmp"
	if err := os.MkdirAll(tmpFolder, os.ModePerm); err != nil {
		fmt.Printf("Error creating tmp directory: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create temporary directory"})
		return
	}

	// Generate a unique filename using a Unix timestamp to prevent file collisions during concurrent requests
	tempFilename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	tempPath := filepath.Join(tmpFolder, tempFilename)

	// 3. Save the uploaded file to the local disk
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		// Log the exact system error to the terminal for debugging
		fmt.Printf("System error while saving file: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the file locally"})
		return
	}

	// 4. Cleanup: Automatically delete the local file after the request finishes (regardless of success or failure)
	defer os.Remove(tempPath)

	// ==========================================
	// AI PIPELINE START
	// ==========================================

	// Step 1: Speech-to-Text (OpenAI Whisper)
	transcript, err := services.TranscribeAudio(tempPath)
	if err != nil {
		fmt.Println("Whisper API Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transcribe audio"})
		return
	}

	// Step 2: Unstructured Text to Structured JSON (OpenAI GPT-4o)
	structuredData, err := services.ExtractData(transcript)
	if err != nil {
		fmt.Println("GPT-4o API Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract structured data from transcript"})
		return
	}

	// ==========================================
	// AI PIPELINE END
	// ==========================================

	// Step 3: Push the "Trinity" (Notes, Tasks, Properties) to HubSpot

	// 3A: Create the Note (Using the dynamically generated Title and Summary)
	htmlBody := fmt.Sprintf("<h2>%s</h2><p>%s</p>", structuredData.Title, structuredData.Summary)
	if err := services.CreateNote(htmlBody); err != nil {
		// We log the error but continue execution so tasks and properties might still succeed
		fmt.Println("HubSpot Note Error:", err)
	}

	// 3B: Create the Tasks (Action items with calculated due dates)
	for _, task := range structuredData.Tasks {
		if err := services.CreateTask(task); err != nil {
			fmt.Println("HubSpot Task Error:", err)
		}
	}

	// 3C: Patch the Contact Properties (Only if the AI extracted relevant data)
	if structuredData.ContactUpdates.LeadStatus != "" {
		if err := services.UpdateContact(structuredData.ContactUpdates); err != nil {
			fmt.Println("HubSpot Contact Update Error:", err)
		}
	}

	// Final Success Response
	c.JSON(http.StatusOK, gin.H{
		"message": "Success! Audio processed. Note, tasks, and contact updates have been saved to HubSpot.",
	})
}
