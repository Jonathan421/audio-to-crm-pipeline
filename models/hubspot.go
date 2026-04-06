package models

// ==========================================
// 1. Models for Notes
// ==========================================

type HubSpotNoteRequest struct {
	Properties   NoteProperties `json:"properties"`
	Associations []Association  `json:"associations,omitempty"`
}

type NoteProperties struct {
	Timestamp string `json:"hs_timestamp"`
	Body      string `json:"hs_note_body"`
}

// ==========================================
// 2. Shared Models for Associations
// These are used to link objects (e.g., attaching Notes or Tasks to a Contact)
// ==========================================

type Association struct {
	To    AssociationTo `json:"to"`
	Types []AssocType   `json:"types"`
}

type AssociationTo struct {
	ID string `json:"id"`
}

type AssocType struct {
	AssociationCategory string `json:"associationCategory"`
	AssociationTypeId   int    `json:"associationTypeId"`
}

// ==========================================
// 3. Models for Tasks
// ==========================================

type HubSpotTaskRequest struct {
	Properties   TaskProperties `json:"properties"`
	Associations []Association  `json:"associations,omitempty"`
}

type TaskProperties struct {
	Timestamp string `json:"hs_timestamp"` // Required by HubSpot to set the creation/due date
	Subject   string `json:"hs_task_subject"`
	Status    string `json:"hs_task_status"`
	Body      string `json:"hs_task_body"`
}

// ==========================================
// 4. Models for CRM Property Updates (Contact Patch)
// ==========================================

type HubSpotContactPatch struct {
	Properties map[string]string `json:"properties"`
}
