package models

// WhisperResponse fängt das Ergebnis der Audio-Transkription ab
type WhisperResponse struct {
	Text string `json:"text"`
}

// OpenAIResponse ist unser Ziel-Format für GPT-4o (Zusammenfassung & To-dos)
type OpenAIResponse struct {
	Summary string   `json:"summary"`
	Todos   []string `json:"todos"`
}
