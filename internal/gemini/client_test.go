package gemini

import "testing"

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Clean JSON",
			input:    `{"suggestions":[]}`,
			expected: `{"suggestions":[]}`,
		},
		{
			name:     "Markdown Wrapped",
			input:    "```json\n{\n  \"suggestions\": []\n}\n```",
			expected: "{\n  \"suggestions\": []\n}",
		},
		{
			name:     "Reasoning Cluttered",
			input:    "<think>Thinking process:\n1. Check status.\n</think>\n{\n  \"suggestions\": []\n}",
			expected: "{\n  \"suggestions\": []\n}",
		},
		{
			name:     "Text Prefix and Suffix",
			input:    "Some text before\n{\n  \"suggestions\": []\n}\nSome text after",
			expected: "{\n  \"suggestions\": []\n}",
		},
		{
			name:     "No brackets",
			input:    "malformed text",
			expected: "malformed text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSON(tt.input)
			if got != tt.expected {
				t.Errorf("extractJSON(%q) = %q; expected %q", tt.input, got, tt.expected)
			}
		})
	}
}
