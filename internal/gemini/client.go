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

// Groq structures
type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqResponseFormat struct {
	Type string `json:"type"`
}

type GroqRequestPayload struct {
	Model          string              `json:"model"`
	Messages       []GroqMessage       `json:"messages"`
	ResponseFormat *GroqResponseFormat `json:"response_format,omitempty"`
	MaxTokens      int                 `json:"max_tokens,omitempty"`
}

type GroqResponsePayload struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func GetSuggestions(apiKey string, model string, apiEndpoint string, cmd string, exitCode int, lastSuggestion string, lastExit int) ([]Suggestion, error) {
	if strings.HasPrefix(apiKey, "gsk_") {
		return getGroqSuggestions(apiKey, model, apiEndpoint, cmd, exitCode, lastSuggestion, lastExit)
	}
	return getGeminiSuggestions(apiKey, model, apiEndpoint, cmd, exitCode, lastSuggestion, lastExit)
}

func getGeminiSuggestions(apiKey string, model string, apiEndpoint string, cmd string, exitCode int, lastSuggestion string, lastExit int) ([]Suggestion, error) {
	pwd, _ := os.Getwd()

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

	sysInstrText := `You are 'frick', a CLI helper tool that fixes mistyped terminal commands.
Your job is to analyze a failed or typoed terminal command and suggest up to 3 corrected command alternatives.

For each suggestion, determine its safety level:
- SAFE: read-only, informational, or harmless commands (e.g., 'git status', 'ls', 'pwd', 'cat').
- WARNING: modifying commands (e.g., 'git commit', 'git checkout', 'docker run', 'mv', 'cp', 'npm install').
- DANGER: highly destructive, irreversible, or superuser commands (e.g., 'rm', 'dd', 'mkfs', 'sudo', 'git reset --hard').

Ensure you return a valid JSON object matching the requested schema.`

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

	schema := &ResponseSchema{
		Type:     "OBJECT",
		Required: []string{"suggestions"},
		Properties: map[string]SchemaProperty{
			"suggestions": {
				Type: "ARRAY",
				Items: &SchemaProperty{
					Type:     "OBJECT",
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

	baseUrl := "https://generativelanguage.googleapis.com"
	if apiEndpoint != "" {
		baseUrl = strings.TrimSuffix(apiEndpoint, "/")
	}
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", baseUrl, model, apiKey)
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

func getGroqSuggestions(apiKey string, model string, apiEndpoint string, cmd string, exitCode int, lastSuggestion string, lastExit int) ([]Suggestion, error) {
	if model == "" || model == "gemma-4-31b-it" {
		model = "qwen/qwen3.6-27b"
	}

	pwd, _ := os.Getwd()

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

	sysInstrText := `You are 'frick', a CLI helper tool that fixes mistyped terminal commands.
Your job is to analyze a failed or typoed terminal command and suggest up to 3 corrected command alternatives.

For each suggestion, determine its safety level:
- SAFE: read-only, informational, or harmless commands (e.g., 'git status', 'ls', 'pwd', 'cat').
- WARNING: modifying commands (e.g., 'git commit', 'git checkout', 'docker run', 'mv', 'cp', 'npm install').
- DANGER: highly destructive, irreversible, or superuser commands (e.g., 'rm', 'dd', 'mkfs', 'sudo', 'git reset --hard').

Keep your thinking and reasoning process extremely short and concise to avoid exceeding token limits.
Respond ONLY with a valid JSON object matching this schema. Do not wrap the response in markdown code blocks. Output raw JSON only:
{
  "suggestions": [
    {
      "command": "corrected command here",
      "safety_level": "SAFE" | "WARNING" | "DANGER"
    }
  ]
}`

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

	payload := GroqRequestPayload{
		Model: model,
		Messages: []GroqMessage{
			{
				Role:    "system",
				Content: sysInstrText,
			},
			{
				Role:    "user",
				Content: promptText,
			},
		},
		ResponseFormat: &GroqResponseFormat{
			Type: "text",
		},
		MaxTokens: 4096,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	baseUrl := "https://api.groq.com"
	if apiEndpoint != "" {
		baseUrl = strings.TrimSuffix(apiEndpoint, "/")
	}
	url := fmt.Sprintf("%s/openai/v1/chat/completions", baseUrl)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Groq request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Groq API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var groqResp GroqResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return nil, fmt.Errorf("failed to decode Groq response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("empty choices returned from Groq API")
	}

	responseText := groqResp.Choices[0].Message.Content

	// Extract JSON block in case Qwen prints thoughts or markdown blocks
	jsonStr := extractJSON(responseText)

	var finalResult APIResponse
	if err := json.Unmarshal([]byte(jsonStr), &finalResult); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from Groq response content: %w. Extracted was: %s, Full was: %s", err, jsonStr, responseText)
	}

	return finalResult.Suggestions, nil
}

func extractJSON(s string) string {
	start := strings.Index(s, "{")
	if start == -1 {
		return s
	}
	end := strings.LastIndex(s, "}")
	if end == -1 || end < start {
		return s
	}
	return s[start : end+1]
}
