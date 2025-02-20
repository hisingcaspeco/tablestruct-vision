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
Now analyze this new layout. Identify the following objects:
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
  itemType: TableMapItemType; // set WALL to 0, DOOR to 1, WINDOW to 2, TABLE_RECT to 3, TABLE_ELLIPSE to 4, LABEL to 5, VIRTUAL to 6
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

if WALL - positionX + positionY = koordinat där väggen börjar.
if WALL - sizeX + sizeY = koordinat där väggen slutar.

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
			/*{Role: "user", Content: []map[string]any{
				{
					"type": "image_url",
					"image_url": map[string]string{
						"url": "https://yetric.se/floor.jpg",
					}},
				{
					"type": "text",
					"text": "In this image I have annotated the tables, chairs, walls, and text labels. Walls are marked with red lines, Tables are marked with blue marking, chairs are marked with green marking, and text labels are marked with yellow marking. Please identify these objects in the image.",
				},

			},
			},*/
			{Role: "user", Content: []map[string]any{
				{
					"type": "image_url",
					"image_url": map[string]string{
						"url": "https://yetric.se/hattmakarn.jpg",
					}},
				{
					"type": "text",
					"text": "In this image I have annotated the tables, chairs, walls, and text labels. Walls are marked with red lines, Tables are marked with blue marking, chairs are marked with green marking, and text labels are marked with yellow marking. Please identify these objects in the image." +
						"Expected JSON-output:" +
						"```json" +
						"{\n  \"objects\": [\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"WALL\",\n      \"positionX\": 0,\n      \"positionY\": 0,\n      \"sizeX\": 1263,\n      \"sizeY\": 768,\n      \"id\": 1000,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"LABEL\",\n      \"positionX\": 650,\n      \"positionY\": 100,\n      \"sizeX\": 100,\n      \"sizeY\": 30,\n      \"tableName\": \"Entré\",\n      \"id\": 1001,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"LABEL\",\n      \"positionX\": 300,\n      \"positionY\": 400,\n      \"sizeX\": 100,\n      \"sizeY\": 30,\n      \"tableName\": \"Lounge\",\n      \"id\": 1002,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"TABLE_RECT\",\n      \"positionX\": 150,\n      \"positionY\": 200,\n      \"sizeX\": 60,\n      \"sizeY\": 60,\n      \"tableName\": \"Table 1\",\n      \"tableNumber\": 1,\n      \"articleId\": 101,\n      \"articleGroupId\": 10,\n      \"includedInResourcePool\": true,\n      \"stockBalance\": 1,\n      \"departmentId\": \"A\",\n      \"sortOrder\": 1,\n      \"chairs\": 4,\n      \"chairsMax\": 6,\n      \"sectionId\": 1,\n      \"webBookable\": true,\n      \"isResourcePool\": false,\n      \"unitId\": 1,\n      \"articleIds\": \"101,102\",\n      \"guestsMin\": 2,\n      \"priority\": 1,\n      \"rotate\": 0,\n      \"id\": 2001,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"TABLE_RECT\",\n      \"positionX\": 250,\n      \"positionY\": 200,\n      \"sizeX\": 60,\n      \"sizeY\": 60,\n      \"tableName\": \"Table 2\",\n      \"tableNumber\": 2,\n      \"articleId\": 102,\n      \"articleGroupId\": 10,\n      \"includedInResourcePool\": true,\n      \"stockBalance\": 1,\n      \"departmentId\": \"A\",\n      \"sortOrder\": 2,\n      \"chairs\": 4,\n      \"chairsMax\": 6,\n      \"sectionId\": 1,\n      \"webBookable\": true,\n      \"isResourcePool\": false,\n      \"unitId\": 1,\n      \"articleIds\": \"101,102\",\n      \"guestsMin\": 2,\n      \"priority\": 1,\n      \"rotate\": 0,\n      \"id\": 2002,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"TABLE_RECT\",\n      \"positionX\": 350,\n      \"positionY\": 200,\n      \"sizeX\": 60,\n      \"sizeY\": 60,\n      \"tableName\": \"Table 3\",\n      \"tableNumber\": 3,\n      \"articleId\": 103,\n      \"articleGroupId\": 10,\n      \"includedInResourcePool\": true,\n      \"stockBalance\": 1,\n      \"departmentId\": \"A\",\n      \"sortOrder\": 3,\n      \"chairs\": 4,\n      \"chairsMax\": 6,\n      \"sectionId\": 1,\n      \"webBookable\": true,\n      \"isResourcePool\": false,\n      \"unitId\": 1,\n      \"articleIds\": \"101,102\",\n      \"guestsMin\": 2,\n      \"priority\": 1,\n      \"rotate\": 0,\n      \"id\": 2003,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"TABLE_RECT\",\n      \"positionX\": 450,\n      \"positionY\": 200,\n      \"sizeX\": 60,\n      \"sizeY\": 60,\n      \"tableName\": \"Table 4\",\n      \"tableNumber\": 4,\n      \"articleId\": 104,\n      \"articleGroupId\": 10,\n      \"includedInResourcePool\": true,\n      \"stockBalance\": 1,\n      \"departmentId\": \"A\",\n      \"sortOrder\": 4,\n      \"chairs\": 4,\n      \"chairsMax\": 6,\n      \"sectionId\": 1,\n      \"webBookable\": true,\n      \"isResourcePool\": false,\n      \"unitId\": 1,\n      \"articleIds\": \"101,102\",\n      \"guestsMin\": 2,\n      \"priority\": 1,\n      \"rotate\": 0,\n      \"id\": 2004,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"TABLE_RECT\",\n      \"positionX\": 650,\n      \"positionY\": 200,\n      \"sizeX\": 60,\n      \"sizeY\": 60,\n      \"tableName\": \"Table 42\",\n      \"tableNumber\": 42,\n      \"articleId\": 142,\n      \"articleGroupId\": 10,\n      \"includedInResourcePool\": true,\n      \"stockBalance\": 1,\n      \"departmentId\": \"A\",\n      \"sortOrder\": 42,\n      \"chairs\": 4,\n      \"chairsMax\": 6,\n      \"sectionId\": 1,\n      \"webBookable\": true,\n      \"isResourcePool\": false,\n      \"unitId\": 1,\n      \"articleIds\": \"101,102\",\n      \"guestsMin\": 2,\n      \"priority\": 1,\n      \"rotate\": 0,\n      \"id\": 2042,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"TABLE_RECT\",\n      \"positionX\": 750,\n      \"positionY\": 200,\n      \"sizeX\": 60,\n      \"sizeY\": 60,\n      \"tableName\": \"Table 43\",\n      \"tableNumber\": 43,\n      \"articleId\": 143,\n      \"articleGroupId\": 10,\n      \"includedInResourcePool\": true,\n      \"stockBalance\": 1,\n      \"departmentId\": \"A\",\n      \"sortOrder\": 43,\n      \"chairs\": 4,\n      \"chairsMax\": 6,\n      \"sectionId\": 1,\n      \"webBookable\": true,\n      \"isResourcePool\": false,\n      \"unitId\": 1,\n      \"articleIds\": \"101,102\",\n      \"guestsMin\": 2,\n      \"priority\": 1,\n      \"rotate\": 0,\n      \"id\": 2043,\n      \"doRemove\": false\n    }\n  ]\n}\n",
				},
			},
			},
			{Role: "user", Content: []map[string]any{
				{
					"type": "image_url",
					"image_url": map[string]string{
						"url": "https://yetric.se/hh.jpg",
					}},
				{
					"type": "text",
					"text": "In this image I have annotated the tables, chairs, walls, and text labels. Walls are marked with red lines, Tables are marked with blue marking, chairs are marked with green marking, and text labels are marked with yellow marking. Please identify these objects in the image." +
						"Expected JSON-output:" +
						"```json" +
						"{\n  \"objects\": [\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"WALL\",\n      \"positionX\": 0,\n      \"positionY\": 0,\n      \"sizeX\": 1263,\n      \"sizeY\": 768,\n      \"id\": 1000,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"LABEL\",\n      \"positionX\": 650,\n      \"positionY\": 100,\n      \"sizeX\": 100,\n      \"sizeY\": 30,\n      \"tableName\": \"Entré\",\n      \"id\": 1001,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"LABEL\",\n      \"positionX\": 300,\n      \"positionY\": 400,\n      \"sizeX\": 100,\n      \"sizeY\": 30,\n      \"tableName\": \"Lounge\",\n      \"id\": 1002,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"TABLE_RECT\",\n      \"positionX\": 150,\n      \"positionY\": 200,\n      \"sizeX\": 60,\n      \"sizeY\": 60,\n      \"tableName\": \"Table 1\",\n      \"tableNumber\": 1,\n      \"articleId\": 101,\n      \"articleGroupId\": 10,\n      \"includedInResourcePool\": true,\n      \"stockBalance\": 1,\n      \"departmentId\": \"A\",\n      \"sortOrder\": 1,\n      \"chairs\": 4,\n      \"chairsMax\": 6,\n      \"sectionId\": 1,\n      \"webBookable\": true,\n      \"isResourcePool\": false,\n      \"unitId\": 1,\n      \"articleIds\": \"101,102\",\n      \"guestsMin\": 2,\n      \"priority\": 1,\n      \"rotate\": 0,\n      \"id\": 2001,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"TABLE_RECT\",\n      \"positionX\": 250,\n      \"positionY\": 200,\n      \"sizeX\": 60,\n      \"sizeY\": 60,\n      \"tableName\": \"Table 2\",\n      \"tableNumber\": 2,\n      \"articleId\": 102,\n      \"articleGroupId\": 10,\n      \"includedInResourcePool\": true,\n      \"stockBalance\": 1,\n      \"departmentId\": \"A\",\n      \"sortOrder\": 2,\n      \"chairs\": 4,\n      \"chairsMax\": 6,\n      \"sectionId\": 1,\n      \"webBookable\": true,\n      \"isResourcePool\": false,\n      \"unitId\": 1,\n      \"articleIds\": \"101,102\",\n      \"guestsMin\": 2,\n      \"priority\": 1,\n      \"rotate\": 0,\n      \"id\": 2002,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"TABLE_RECT\",\n      \"positionX\": 350,\n      \"positionY\": 200,\n      \"sizeX\": 60,\n      \"sizeY\": 60,\n      \"tableName\": \"Table 3\",\n      \"tableNumber\": 3,\n      \"articleId\": 103,\n      \"articleGroupId\": 10,\n      \"includedInResourcePool\": true,\n      \"stockBalance\": 1,\n      \"departmentId\": \"A\",\n      \"sortOrder\": 3,\n      \"chairs\": 4,\n      \"chairsMax\": 6,\n      \"sectionId\": 1,\n      \"webBookable\": true,\n      \"isResourcePool\": false,\n      \"unitId\": 1,\n      \"articleIds\": \"101,102\",\n      \"guestsMin\": 2,\n      \"priority\": 1,\n      \"rotate\": 0,\n      \"id\": 2003,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"TABLE_RECT\",\n      \"positionX\": 450,\n      \"positionY\": 200,\n      \"sizeX\": 60,\n      \"sizeY\": 60,\n      \"tableName\": \"Table 4\",\n      \"tableNumber\": 4,\n      \"articleId\": 104,\n      \"articleGroupId\": 10,\n      \"includedInResourcePool\": true,\n      \"stockBalance\": 1,\n      \"departmentId\": \"A\",\n      \"sortOrder\": 4,\n      \"chairs\": 4,\n      \"chairsMax\": 6,\n      \"sectionId\": 1,\n      \"webBookable\": true,\n      \"isResourcePool\": false,\n      \"unitId\": 1,\n      \"articleIds\": \"101,102\",\n      \"guestsMin\": 2,\n      \"priority\": 1,\n      \"rotate\": 0,\n      \"id\": 2004,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"TABLE_RECT\",\n      \"positionX\": 650,\n      \"positionY\": 200,\n      \"sizeX\": 60,\n      \"sizeY\": 60,\n      \"tableName\": \"Table 42\",\n      \"tableNumber\": 42,\n      \"articleId\": 142,\n      \"articleGroupId\": 10,\n      \"includedInResourcePool\": true,\n      \"stockBalance\": 1,\n      \"departmentId\": \"A\",\n      \"sortOrder\": 42,\n      \"chairs\": 4,\n      \"chairsMax\": 6,\n      \"sectionId\": 1,\n      \"webBookable\": true,\n      \"isResourcePool\": false,\n      \"unitId\": 1,\n      \"articleIds\": \"101,102\",\n      \"guestsMin\": 2,\n      \"priority\": 1,\n      \"rotate\": 0,\n      \"id\": 2042,\n      \"doRemove\": false\n    },\n    {\n      \"templateId\": 1,\n      \"validFrom\": \"2025-02-20T00:00:00Z\",\n      \"itemType\": \"TABLE_RECT\",\n      \"positionX\": 750,\n      \"positionY\": 200,\n      \"sizeX\": 60,\n      \"sizeY\": 60,\n      \"tableName\": \"Table 43\",\n      \"tableNumber\": 43,\n      \"articleId\": 143,\n      \"articleGroupId\": 10,\n      \"includedInResourcePool\": true,\n      \"stockBalance\": 1,\n      \"departmentId\": \"A\",\n      \"sortOrder\": 43,\n      \"chairs\": 4,\n      \"chairsMax\": 6,\n      \"sectionId\": 1,\n      \"webBookable\": true,\n      \"isResourcePool\": false,\n      \"unitId\": 1,\n      \"articleIds\": \"101,102\",\n      \"guestsMin\": 2,\n      \"priority\": 1,\n      \"rotate\": 0,\n      \"id\": 2043,\n      \"doRemove\": false\n    }\n  ]\n}\n" +,
				},
			},
			},
			/*{Role: "user", Content: []map[string]any{
				{
					"type": "image_url",
					"image_url": map[string]string{
						"url": "https://yetric.se/matsal.jpg",
					}},
				{
					"type": "text",
					"text": "In this image I have annotated the tables, chairs, walls, and text labels. Walls are marked with red lines, Tables are marked with blue marking, chairs are marked with green marking, and text labels are marked with yellow marking. Please identify these objects in the image.",
				},
			},
			},*/
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
