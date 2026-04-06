package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"voiceline-mvp/models" // Passe den Modulnamen an, falls abweichend
)

// TranscribeAudio nimmt den Dateipfad und schickt die Datei an die Whisper API
func TranscribeAudio(filePath string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY ist nicht gesetzt")
	}

	// 1. Die lokale Datei öffnen
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("fehler beim Öffnen der Audiodatei: %w", err)
	}
	defer file.Close()

	// 2. Einen Multipart-Body aufbauen (wie in Postman)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Das Feld für die Datei hinzufügen ("file" ist der von OpenAI erwartete Name)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("fehler beim Erstellen des FormFiles: %w", err)
	}
	io.Copy(part, file)

	// Das Feld für das Modell hinzufügen (wir nutzen "whisper-1")
	writer.WriteField("model", "whisper-1")

	// Wichtig: Den Writer schließen, damit der abschließende Boundary geschrieben wird
	writer.Close()

	// 3. Den HTTP-Request an OpenAI vorbereiten
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("fehler beim Erstellen des Requests: %w", err)
	}

	// Header setzen (Auth und Content-Type inklusive der generierten Boundary)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 4. Request absenden
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fehler beim API-Call zu OpenAI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fehler auslesen, falls OpenAI meckert
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai fehler (Status %d): %s", resp.StatusCode, string(respBody))
	}

	// 5. Antwort in unser Struct entpacken
	var whisperResp models.WhisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&whisperResp); err != nil {
		return "", fmt.Errorf("fehler beim Dekodieren der Antwort: %w", err)
	}

	return whisperResp.Text, nil

}

// ... (bestehender Code von TranscribeAudio) ...

// Hilfs-Structs für den Request an GPT-4o (nur intern im Service genutzt)
type chatRequest struct {
	Model          string         `json:"model"`
	ResponseFormat map[string]any `json:"response_format"`
	Messages       []chatMessage  `json:"messages"`
	Temperature    float32        `json:"temperature"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ExtractData nimmt den Rohtext und extrahiert Zusammenfassung und To-dos via GPT-4o
func ExtractData(transcript string) (models.OpenAIResponse, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	var result models.OpenAIResponse

	systemPrompt := `Du bist ein hochpräziser KI-Assistent für Vertriebsmitarbeiter im Außendienst. 
Deine Aufgabe ist es, das unstrukturierte Transkript einer Sprachnotiz nach einem Kundentermin zu analysieren und die wichtigsten Informationen zu extrahieren.
Analysiere den Text und extrahiere:
1. Eine professionelle, kompakte Zusammenfassung.
2. Eine präzise Liste aller To-dos.

WICHTIG: Antworte AUSSCHLIESSLICH mit einem validen JSON-Objekt.
Struktur: {"summary": "...", "todos": ["...", "..."]}`

	// 1. Request-Payload zusammenbauen
	reqBody := chatRequest{
		Model: "gpt-4o",
		ResponseFormat: map[string]any{
			"type": "json_object", // Zwingt die KI, nur JSON zurückzugeben
		},
		Temperature: 0.2, // Niedrige Temperatur für sachliche Extraktion
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: transcript},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return result, fmt.Errorf("fehler beim Erstellen des JSON-Bodys: %w", err)
	}

	// 2. HTTP Request an OpenAI Chat Completions senden
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return result, fmt.Errorf("fehler beim Erstellen des Requests: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("fehler beim API-Call zu OpenAI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("openai fehler: Status %d", resp.StatusCode)
	}

	// 3. Antwort dekodieren
	var apiResponse chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return result, fmt.Errorf("fehler beim Dekodieren der Antwort: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return result, fmt.Errorf("openai hat keine Antwort generiert")
	}

	// 4. Den JSON-String, den die KI generiert hat, in unser finales Struct umwandeln
	contentString := apiResponse.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(contentString), &result); err != nil {
		return result, fmt.Errorf("fehler beim Entpacken des KI-JSONs: %w", err)
	}

	return result, nil
}
