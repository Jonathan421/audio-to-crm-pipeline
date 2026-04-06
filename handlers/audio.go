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

func HandleAudioUpload(c *gin.Context) {
	// 1. Die Datei extrahieren
	file, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keine Audiodatei im Feld 'audio' gefunden"})
		return
	}

	// 2. NEU: Einen lokalen tmp-Ordner nutzen und erstellen, falls er fehlt
	tmpFolder := "./tmp"
	if err := os.MkdirAll(tmpFolder, os.ModePerm); err != nil {
		fmt.Printf("Fehler beim Erstellen des Ordners: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Konnte tmp-Ordner nicht anlegen"})
		return
	}

	tempFilename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	tempPath := filepath.Join(tmpFolder, tempFilename)

	// 3. Die Datei lokal speichern
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		// NEU: Wir loggen den genauen Systemfehler in dein Terminal!
		fmt.Printf("Systemfehler beim Speichern: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Speichern der Datei"})
		return
	}

	// 4. Cleanup (löscht die Datei nach dem Request)
	defer os.Remove(tempPath)

	// ==========================================
	// KI-PIPELINE START
	// ==========================================

	// Schritt 1: Audio zu Text (Whisper)
	transcript, err := services.TranscribeAudio(tempPath)
	if err != nil {
		fmt.Println("Whisper Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler bei der Transkription"})
		return
	}

	// Schritt 2: Text strukturieren (GPT-4o)
	structuredData, err := services.ExtractData(transcript)
	if err != nil {
		fmt.Println("GPT-4o Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler bei der KI-Extraktion"})
		return
	}

	// ==========================================
	// KI-PIPELINE ENDE
	// ==========================================

	// Schritt 3: HTML für HubSpot generieren
	htmlBody := fmt.Sprintf("<h2>Voiceline Termin-Zusammenfassung</h2><p>%s</p><h3>Aktionspunkte (To-Dos)</h3><ul>", structuredData.Summary)
	for _, todo := range structuredData.Todos {
		htmlBody += fmt.Sprintf("<li>[ ] %s</li>", todo)
	}
	htmlBody += "</ul>"

	// Schritt 4: Zu HubSpot senden
	err = services.CreateNote(htmlBody)
	if err != nil {
		fmt.Println("HubSpot Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Fehler beim Sync mit HubSpot"})
		return
	}

	// Finale Erfolgsmeldung
	c.JSON(http.StatusOK, gin.H{
		"message": "Erfolg! Audio analysiert und als Notiz in HubSpot gespeichert.",
	})
}
