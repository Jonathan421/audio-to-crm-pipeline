package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"voiceline-mvp/models" // Passe den Import-Pfad ggf. an
)

// CreateNote sendet die formatierte Notiz an HubSpot
func CreateNote(htmlBody string) error {
	token := os.Getenv("HUBSPOT_TOKEN")
	contactID := os.Getenv("HUBSPOT_CONTACT_ID")

	if token == "" || contactID == "" {
		return fmt.Errorf("HUBSPOT_TOKEN oder HUBSPOT_CONTACT_ID fehlt in der .env")
	}

	// 1. Das JSON für HubSpot zusammenbauen
	reqBody := models.HubSpotNoteRequest{
		Properties: models.NoteProperties{
			Timestamp: fmt.Sprintf("%d", time.Now().UnixMilli()),
			Body:      htmlBody,
		},
		Associations: []models.Association{
			{
				To: models.AssociationTo{ID: contactID},
				Types: []models.AssocType{
					{
						AssociationCategory: "HUBSPOT_DEFINED",
						AssociationTypeId:   202, // ID für Note-to-Contact
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("fehler beim JSON marshaling: %w", err)
	}

	// 2. Request an HubSpot senden
	url := "https://api.hubapi.com/crm/v3/objects/notes"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("fehler beim Erstellen des Requests: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fehler beim HubSpot API Call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hubspot API fehler, Statuscode: %d", resp.StatusCode)
	}

	return nil
}
