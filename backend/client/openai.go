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
- Walls
- Doors
- Text labels

The response should map to this type of JSON structure:

export enum TableMapItemType {
  WALL,
  DOOR,
  WINDOW,
  TABLE_RECT,
  TABLE_ELLIPSE,
  LABEL,
  VIRTUAL,
}

export type TableRecord = {
  templateId: number;
  validFrom: string; // ISO 8601 date string
  itemType: TableMapItemType;
  positionX: number;
  positionY: number;
  sizeX: number;
  sizeY: number;
  tableName: string;
  tableNumber: number;
  articleId: number;
  articleGroupId: number;
  includedInResourcePool: boolean;
  stockBalance: number;
  departmentId?: string;
  sortOrder: number;
  chairs: number;
  chairsMax: number;
  sectionId: number;
  webBookable: boolean;
  isResourcePool: boolean;
  unitId: number;
  articleIds: string;
  guestsMin: number;
  priority: number;
  rotate: number;
  id: number;
  doRemove: boolean;
};

Return the response as structured JSON:
{
  "objects": [
    // list of TableRecord objects
  ]
}`

	prompt += "\n\n!!!!!!!\n\n🚨 IMPORTANT: Do NOT include \\`json or \\` in your response.\nOnly return the JSON output as raw JSON, with no extra formatting."

	// Construct OpenAI request
	requestBody := OpenAIRequest{
		Model: "gpt-4o", // Ensure we're using a vision-capable model
		Messages: []Message{
			{Role: "system", Content: "You are an expert in analyzing restaurant layouts. You will get a blueprint or a sketch of a restaurant and need to identify tables, walls, doors, and text labels. You will get 4 examples to learn from."},
			{Role: "user", Content: []map[string]any{
				{
					"type": "image_url",
					"image_url": map[string]string{
						"url": "https://yetric.se/hackathon/floor.jpg",
					}},
				{
					"type": "text",
					"text": "In this image I have annotated the tables, chairs, walls, and text labels. Walls are marked with red lines, Tables are marked with blue marking, chairs are marked with green marking, and text labels are marked with yellow marking. Please identify these objects in the image.",
				},
			},
			},
			{Role: "user", Content: []map[string]any{
				{
					"type": "image_url",
					"image_url": map[string]string{
						"url": "https://yetric.se/hackathon/hattmakarn.jpg",
					}},
				{
					"type": "text",
					"text": "In this image I have annotated the tables, chairs, walls, and text labels. Walls are marked with red lines, Tables are marked with blue marking, chairs are marked with green marking, and text labels are marked with yellow marking. Please identify these objects in the image.",
				},
			},
			},
			{Role: "user", Content: []map[string]any{
				{
					"type": "image_url",
					"image_url": map[string]string{
						"url": "https://yetric.se/hackathon/hh.jpg",
					}},
				{
					"type": "text",
					"text": "In this image I have annotated the tables, chairs, walls, and text labels. Walls are marked with red lines, Tables are marked with blue marking, chairs are marked with green marking, and text labels are marked with yellow marking. Please identify these objects in the image.",
				},
			},
			},
			{Role: "user", Content: []map[string]any{
				{
					"type": "image_url",
					"image_url": map[string]string{
						"url": "https://yetric.se/hackathon/matsal.jpg",
					}},
				{
					"type": "text",
					"text": "In this image I have annotated the tables, chairs, walls, and text labels. Walls are marked with red lines, Tables are marked with blue marking, chairs are marked with green marking, and text labels are marked with yellow marking. Please identify these objects in the image.",
				},
			},
			},
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
