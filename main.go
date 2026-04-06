package main

import (
	"log"
	"os"
	"voiceline-mvp/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Umgebungsvariablen aus der .env Datei laden
	err := godotenv.Load()
	if err != nil {
		log.Println("Warnung: Keine .env Datei gefunden, nutze System-Umgebungsvariablen")
	}

	// 2. Prüfen, ob die wichtigsten Keys da sind (Fail Fast)
	if os.Getenv("OPENAI_API_KEY") == "" || os.Getenv("HUBSPOT_TOKEN") == "" {
		log.Fatal("KRITISCHER FEHLER: OPENAI_API_KEY oder HUBSPOT_TOKEN fehlen in der .env")
	}

	// 3. Gin Router im Standard-Modus initialisieren
	router := gin.Default()

	// 4. Einen simplen Health-Check Endpunkt (gut fürs Debugging)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong - Server läuft!",
		})
	})

	// 5. Hier kommt später unsere Haupt-Route für den Audio-Upload hin
	// router.POST("/api/v1/note", handlers.HandleAudioUpload)
	router.POST("/api/v1/note", handlers.HandleAudioUpload)

	// 6. Server starten
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server startet auf Port %s...", port)
	router.Run(":" + port)
}
