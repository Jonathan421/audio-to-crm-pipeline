package models

// ==========================================
// 1. Models for Audio Transcription
// ==========================================

// WhisperResponse represents the JSON structure returned by the OpenAI Whisper API.
type WhisperResponse struct {
	Text string `json:"text"`
}

// ==========================================
// 2. Models for Structured AI Extraction
// ==========================================

// Task represents a single actionable item extracted from the transcript.
type Task struct {
	Title   string `json:"title"`
	DueDate string `json:"due_date"`
}

// ContactUpdates holds the specific CRM property values extracted from the conversation.
type ContactUpdates struct {
	LeadStatus     string `json:"lead_status"`
	LifecycleStage string `json:"lifecycle_stage"`
	Revenue        string `json:"revenue"`
	Industry       string `json:"industry"`
}

// OpenAIResponse represents the strict JSON structure we force GPT-4o to return.
// It contains the generated title, summary, action items, and data points for the CRM.
type OpenAIResponse struct {
	Title          string         `json:"title"`
	Summary        string         `json:"summary"`
	Tasks          []Task         `json:"tasks"`
	ContactUpdates ContactUpdates `json:"contact_updates"`
}
