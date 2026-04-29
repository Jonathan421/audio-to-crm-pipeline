# Voiceline MVP: Audio-to-CRM Pipeline

This repository contains a minimal viable product (MVP) for an audio-to-CRM pipeline, built as a tech task for Voiceline. It receives an audio note, transcribes it, extracts structured data (summary and to-dos) using AI, and pushes the formatted result directly into HubSpot.

## 🚀 Tech Stack & Architecture

- **Backend:** Go (Golang) with the Gin framework for fast and reliable HTTP routing.
- **Architecture:** "Clean Architecture Light". Handlers and Services are separated to keep the code modular, testable, and easy to extend.
- **AI / LLM:** - **OpenAI Whisper API (`whisper-1`):** Used for highly accurate speech-to-text transcription.
  - **OpenAI Chat Completions (`gpt-4o`):** Used with strict `json_object` response formatting and a system prompt to extract business value and actionable items.
- **CRM Integration:** HubSpot CRM API v3 (Private App Token).

## 🛠 Prerequisites

To run this project locally, you need:
1. Go installed on your machine (v1.18+).
2. An OpenAI API Key with active billing.
3. A HubSpot Private App Token (with `crm.objects.contacts.read/write` scopes) and a valid Contact ID.

Create a `.env` file in the root directory (you can copy the structure from `.env.example` if provided):

```env
PORT=8080
OPENAI_API_KEY=sk-YourOpenAIKey
HUBSPOT_TOKEN=pat-na1-YourHubSpotToken
HUBSPOT_CONTACT_ID=YourTargetContactID
```

```
curl -X POST http://localhost:8080/api/v1/note -F "audio=@audio_inputs/test_audio.m4a"
```