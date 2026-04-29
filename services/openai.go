package services

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"voiceline-mvp/models"
)

// TranscribeAudio takes a local file path and sends the audio file to the OpenAI Whisper API.
// It returns the transcribed text as a string.
func TranscribeAudio(filePath string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is not set in the environment")
	}

	// 1. Open the local audio file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening audio file: %w", err)
	}
	defer file.Close()

	// 2. Construct a multipart/form-data body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file field ("file" is the exact key expected by OpenAI)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("error creating form file: %w", err)
	}
	io.Copy(part, file)

	// Add the model field
	writer.WriteField("model", "whisper-1")

	// Important: Close the writer before creating the request to ensure the terminating boundary is written
	writer.Close()

	// 3. Prepare the HTTP POST request to OpenAI
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set headers (Auth and Content-Type, which includes the auto-generated boundary)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 4. Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error executing OpenAI API call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the error body if OpenAI rejects the request (e.g., file too large, invalid format)
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai error (Status %d): %s", resp.StatusCode, string(respBody))
	}

	// 5. Decode the successful JSON response into our WhisperResponse struct
	var whisperResp models.WhisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&whisperResp); err != nil {
		return "", fmt.Errorf("error decoding OpenAI response: %w", err)
	}

	return whisperResp.Text, nil
}

// ==========================================
// Helper structs for the GPT-4o API request
// These are used internally in this service and don't need to be exported to the models package
// ==========================================

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

//go:embed system_prompt.txt
var systemPromptTemplate string

// ExtractData takes the raw transcript and uses GPT-4o to extract a summary, tasks, and CRM property updates.
// It forces the AI to return a strict JSON structure mapped to our models.OpenAIResponse.
func ExtractData(transcript string) (models.OpenAIResponse, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	var result models.OpenAIResponse

	// 1. Fetch the current date dynamically.
	// This is crucial to give the LLM an anchor point for calculating relative dates (e.g., "next week").
	today := time.Now().Format("2006-01-02")

	// 2. Inject the dynamic date into the system prompt using fmt.Sprintf
	systemPrompt := fmt.Sprintf(systemPromptTemplate, today, today)

	// 3. Build the request payload
	reqBody := chatRequest{
		Model: "gpt-4o",
		ResponseFormat: map[string]any{
			"type": "json_object", // Force the model into JSON mode to guarantee parseable output
		},
		Temperature: 0.2, // Low temperature ensures factual, deterministic extraction rather than creative writing
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: transcript},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return result, fmt.Errorf("error marshaling JSON body: %w", err)
	}

	// 4. Send the HTTP POST request to the OpenAI Chat Completions endpoint
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return result, fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("error executing OpenAI API call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("openai error: Status %d", resp.StatusCode)
	}

	// 5. Decode the API response wrapper
	var apiResponse chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return result, fmt.Errorf("error decoding API response: %w", err)
	}

	if len(apiResponse.Choices) == 0 {
		return result, fmt.Errorf("openai did not generate a response")
	}

	// 6. Extract the actual JSON string generated by the AI and unmarshal it into our Go struct
	contentString := apiResponse.Choices[0].Message.Content
	if err := json.Unmarshal([]byte(contentString), &result); err != nil {
		return result, fmt.Errorf("error unmarshaling the AI-generated JSON into Go struct: %w", err)
	}

	return result, nil
}
