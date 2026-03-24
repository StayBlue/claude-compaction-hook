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
	"fmt"
	"os"
	"strings"
)

const compactMarker = "Summary provided via SessionStart hook."

const compactInstructions = `# Compact Instructions

Do not summarize the conversation. Output only the following and nothing else:

Summary provided via SessionStart hook.
`

// findClaudeMD finds the CLAUDE.md file in the project root (parent of .claude/).
func findClaudeMD(claudeDir string) string {
	projectRoot := claudeDir[:len(claudeDir)-len("/.claude")]
	if projectRoot == "" {
		projectRoot = "/"
	}
	return projectRoot + "/CLAUDE.md"
}

// findCompactSection locates the "# Compact Instructions" section in content.
// Returns the start index of the heading and end index of the section
// (up to the next heading or EOF). Returns -1, -1 if not found.
func findCompactSection(content string) (int, int) {
	heading := "# Compact Instructions"
	idx := strings.Index(content, heading)
	if idx == -1 {
		return -1, -1
	}

	// Find the end: next heading at same or higher level, or EOF
	rest := content[idx+len(heading):]
	endOffset := len(rest)
	for i := 0; i < len(rest); i++ {
		if rest[i] == '\n' && i+1 < len(rest) && rest[i+1] == '#' {
			endOffset = i + 1
			break
		}
	}

	return idx, idx + len(heading) + endOffset
}

// installCompactInstructions appends compact instructions to CLAUDE.md.
func installCompactInstructions(claudeDir string) (string, error) {
	path := findClaudeMD(claudeDir)

	existing := ""
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	}

	start, end := findCompactSection(existing)
	if start != -1 {
		section := existing[start:end]
		if strings.Contains(section, compactMarker) {
			// Already installed
			return path, nil
		}
		return "", fmt.Errorf("%s already has a Compact Instructions section — remove it first or add %s to it manually", path, compactMarker)
	}

	// Append
	separator := "\n"
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		separator = "\n\n"
	}

	content := existing + separator + compactInstructions

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", path, err)
	}

	return path, nil
}

// uninstallCompactInstructions removes our compact instructions from CLAUDE.md.
func uninstallCompactInstructions(claudeDir string) error {
	path := findClaudeMD(claudeDir)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	content := string(data)
	start, end := findCompactSection(content)
	if start == -1 {
		return nil
	}

	section := content[start:end]
	if !strings.Contains(section, compactMarker) {
		// Not our section
		return nil
	}

	// Remove the section and any preceding blank line
	if start > 0 && content[start-1] == '\n' {
		start--
	}

	newContent := content[:start] + content[end:]
	newContent = strings.TrimRight(newContent, "\n\t ") + "\n"

	return os.WriteFile(path, []byte(newContent), 0o644)
}
