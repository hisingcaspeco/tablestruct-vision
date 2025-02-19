package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAI API URL
const openAIURL = "https://api.openai.com/v1/chat/completions"

// OpenAIClient struct
type OpenAIClient struct {
	APIKey string
}

// NewOpenAIClient initializes the OpenAI client
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{APIKey: apiKey}
}

// OpenAIRequest struct
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message struct
type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// OpenAIResponse struct
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// SendImageToOpenAI detects tables, bars, walls, etc., and returns JSON
func (c *OpenAIClient) SendImageToOpenAI(base64Image string) (string, error) {
	prompt := `
Analyze the provided blueprint, layout, or sketch of a restaurant. Identify the following objects:
- Tables
- Bars
- Walls
- Chairs
- Doors
- Windows

Return the response as structured JSON:
{
  "objects": [
    { "type": "table", "x": 120, "y": 250, "width": 50, "height": 50 },
    { "type": "bar", "x": 300, "y": 400, "width": 100, "height": 50 },
    { "type": "wall", "x": 0, "y": 0, "width": 800, "height": 20 },
    { "type": "chair", "x": 130, "y": 260, "width": 20, "height": 20 }
  ]
}`

	prompt += "\n\n!!!!!!!\n\n🚨 IMPORTANT: Do NOT include \\`json or \\` in your response.\nOnly return the JSON output as raw JSON, with no extra formatting."

	// Construct OpenAI request
	requestBody := OpenAIRequest{
		Model: "gpt-4o", // Ensure we're using a vision-capable model
		Messages: []Message{
			{Role: "system", Content: "You are an expert in analyzing restaurant layouts."},
			{Role: "user", Content: prompt}, // Requesting structured JSON
			{Role: "user", Content: []map[string]any{
				{
					"type": "image_url",
					"image_url": map[string]string{
						"url": fmt.Sprintf("data:image/png;base64,%s", base64Image),
					},
				},
			}},
		},
	}

	// Convert requestBody to JSON
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", openAIURL, bytes.NewBuffer(requestJSON))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Log the full response body for debugging
	fmt.Println("🔍 OpenAI Raw Response:", string(body))

	// Decode OpenAI response
	var openAIResponse OpenAIResponse
	if err := json.Unmarshal(body, &openAIResponse); err != nil {
		return "", err
	}

	// Check if OpenAI returned an error
	if openAIResponse.Error.Message != "" {
		return "", fmt.Errorf("OpenAI API error: %s", openAIResponse.Error.Message)
	}

	// Return OpenAI's structured JSON output
	if len(openAIResponse.Choices) > 0 {
		rawResponse := openAIResponse.Choices[0].Message.Content

		return strings.Trim(rawResponse, "```json```"), nil
	}

	return "", fmt.Errorf("no response from OpenAI (empty choices)")
}
