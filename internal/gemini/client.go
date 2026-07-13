package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Suggestion struct {
	Command     string `json:"command"`
	SafetyLevel string `json:"safety_level"`
}

type APIResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
}

// Request structures for Gemini API
type Part struct {
	Text string `json:"text"`
}

type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type SystemInstruction struct {
	Parts []Part `json:"parts"`
}

type ResponseSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]SchemaProperty `json:"properties"`
	Required   []string                  `json:"required"`
}

type SchemaProperty struct {
	Type        string                    `json:"type"`
	Description string                    `json:"description,omitempty"`
	Items       *SchemaProperty           `json:"items,omitempty"`
	Properties  map[string]SchemaProperty `json:"properties,omitempty"`
	Required    []string                  `json:"required,omitempty"`
	Enum        []string                  `json:"enum,omitempty"`
}

type GenerationConfig struct {
	ResponseMimeType string          `json:"responseMimeType"`
	ResponseSchema   *ResponseSchema `json:"responseSchema,omitempty"`
}

type RequestPayload struct {
	Contents          []Content          `json:"contents"`
	SystemInstruction *SystemInstruction `json:"systemInstruction,omitempty"`
	GenerationConfig  *GenerationConfig  `json:"generationConfig,omitempty"`
}

// Gemini Response structures
type Candidate struct {
	Content struct {
		Parts []Part `json:"parts"`
	} `json:"content"`
}

type ResponsePayload struct {
	Candidates []Candidate `json:"candidates"`
}

func GetSuggestions(apiKey string, model string, cmd string, exitCode int, lastSuggestion string, lastExit int) ([]Suggestion, error) {
	pwd, _ := os.Getwd()

	// Get file list in current directory (up to 30 files)
	var files []string
	entries, err := os.ReadDir(".")
	if err == nil {
		for i, entry := range entries {
			if i >= 30 {
				break
			}
			files = append(files, entry.Name())
		}
	}

	// Environment context
	envInfo := ""
	if ve := os.Getenv("VIRTUAL_ENV"); ve != "" {
		envInfo += fmt.Sprintf("Python virtualenv is active: %s. ", filepath.Base(ve))
	}
	if ce := os.Getenv("CONDA_DEFAULT_ENV"); ce != "" {
		envInfo += fmt.Sprintf("Conda environment is active: %s. ", ce)
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	} else {
		shell = filepath.Base(shell)
	}

	// System instruction
	sysInstrText := `You are 'frick', a CLI helper tool that fixes mistyped terminal commands.
Your job is to analyze a failed or typoed terminal command and suggest up to 3 corrected command alternatives.

For each suggestion, determine its safety level:
- SAFE: read-only, informational, or harmless commands (e.g., 'git status', 'ls', 'pwd', 'cat').
- WARNING: modifying commands (e.g., 'git commit', 'git checkout', 'docker run', 'mv', 'cp', 'npm install').
- DANGER: highly destructive, irreversible, or superuser commands (e.g., 'rm', 'dd', 'mkfs', 'sudo', 'git reset --hard').

Ensure you return a valid JSON object matching the requested schema.`

	// Prompt payload
	promptText := fmt.Sprintf(`Analyze this failed command:
- Command run: %s
- Exit code: %d
- Operating System: %s
- Shell: %s
- Directory: %s`, cmd, exitCode, runtime.GOOS, shell, pwd)

	if len(files) > 0 {
		promptText += fmt.Sprintf("\n- Files in current directory: %s", strings.Join(files, ", "))
	}
	if envInfo != "" {
		promptText += fmt.Sprintf("\n- Environment: %s", envInfo)
	}
	if lastSuggestion != "" && lastExit != 0 {
		promptText += fmt.Sprintf("\n- Double-Frick Warning: The user previously approved your suggested correction '%s' for a failed command, but that corrected command also failed with exit code %d. Do NOT suggest '%s' again. Provide a different correction.", lastSuggestion, lastExit, lastSuggestion)
	}

	// Define response schema
	schema := &ResponseSchema{
		Type: "OBJECT",
		Required: []string{"suggestions"},
		Properties: map[string]SchemaProperty{
			"suggestions": {
				Type: "ARRAY",
				Items: &SchemaProperty{
					Type: "OBJECT",
					Required: []string{"command", "safety_level"},
					Properties: map[string]SchemaProperty{
						"command": {
							Type: "STRING",
						},
						"safety_level": {
							Type: "STRING",
							Enum: []string{"SAFE", "WARNING", "DANGER"},
						},
					},
				},
			},
		},
	}

	payload := RequestPayload{
		Contents: []Content{
			{
				Role: "user",
				Parts: []Part{
					{Text: promptText},
				},
			},
		},
		SystemInstruction: &SystemInstruction{
			Parts: []Part{
				{Text: sysInstrText},
			},
		},
		GenerationConfig: &GenerationConfig{
			ResponseMimeType: "application/json",
			ResponseSchema:   schema,
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp ResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Candidates) == 0 || len(apiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty candidates returned from API")
	}

	responseText := apiResp.Candidates[0].Content.Parts[0].Text

	var finalResult APIResponse
	if err := json.Unmarshal([]byte(responseText), &finalResult); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from response text: %w. Response was: %s", err, responseText)
	}

	return finalResult.Suggestions, nil
}
