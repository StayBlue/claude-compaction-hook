// Copyright 2026 StayBlue
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// transcriptLine represents a line in the Claude Code transcript JSONL.
// The role/content are nested inside a "message" field.
type transcriptLine struct {
	Message transcriptMessage `json:"message"`
}

// transcriptMessage represents the inner message object.
type transcriptMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// contentBlock represents a structured content block within a message.
type contentBlock struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Name  string `json:"name"`
	Input json.RawMessage `json:"input"`
}

// formatTranscript reads a JSONL transcript file and returns a
// human-readable plaintext representation.
func formatTranscript(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open transcript: %w", err)
	}
	defer f.Close()

	var b strings.Builder
	scanner := bufio.NewScanner(f)

	// Increase scanner buffer for large lines (tool results can be big)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	first := true
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var tl transcriptLine
		if err := json.Unmarshal([]byte(line), &tl); err != nil {
			continue // Skip malformed lines
		}
		msg := &tl.Message

		text := extractText(msg)
		if text == "" {
			continue
		}

		if !first {
			b.WriteByte('\n')
		}
		first = false

		fmt.Fprintf(&b, "[%s]: %s\n", msg.Role, text)
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read transcript: %w", err)
	}

	return b.String(), nil
}

// extractText pulls readable text from a message's content field.
// Content can be a plain string or an array of content blocks.
func extractText(msg *transcriptMessage) string {
	if len(msg.Content) == 0 {
		return ""
	}

	// Try as plain string first
	var s string
	if err := json.Unmarshal(msg.Content, &s); err == nil {
		return s
	}

	// Try as array of content blocks
	var blocks []contentBlock
	if err := json.Unmarshal(msg.Content, &blocks); err != nil {
		return ""
	}

	var parts []string
	for _, block := range blocks {
		switch block.Type {
		case "text":
			if block.Text != "" {
				parts = append(parts, block.Text)
			}
		case "tool_use":
			if block.Name != "" {
				parts = append(parts, fmt.Sprintf("[tool: %s]", block.Name))
			}
		case "tool_result":
			// Extract text from tool results if present
			var resultText string
			if err := json.Unmarshal(block.Input, &resultText); err == nil && resultText != "" {
				parts = append(parts, fmt.Sprintf("[result]: %s", resultText))
			}
		}
	}

	return strings.Join(parts, "\n")
}
