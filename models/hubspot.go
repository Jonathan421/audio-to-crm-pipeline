package models

type HubSpotNoteRequest struct {
	Properties   NoteProperties `json:"properties"`
	Associations []Association  `json:"associations,omitempty"`
}

type NoteProperties struct {
	Timestamp string `json:"hs_timestamp"`
	Body      string `json:"hs_note_body"`
}

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
