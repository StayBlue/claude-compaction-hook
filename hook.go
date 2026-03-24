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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// hookInput is the JSON structure Claude Code sends on stdin to hooks.
type hookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Trigger        string `json:"trigger"`
}

// stateDir returns a directory for temp files between hooks.
func stateDir() (string, error) {
	dir := os.TempDir()
	return filepath.Join(dir, "compact-hook"), nil
}

// stateFile returns the path to the state file for a given session.
func stateFile(sessionID string) (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, sessionID), nil
}

// shellQuote wraps a string in single quotes for safe shell use.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// readHookInput reads and parses the JSON hook input from stdin.
func readHookInput() (*hookInput, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	var input hookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("failed to parse hook input: %w", err)
	}

	return &input, nil
}

// debugLog appends a timestamped line to /tmp/compact-hook/debug.log.
func debugLog(format string, args ...any) {
	f, err := os.OpenFile("/tmp/compact-hook/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, format+"\n", args...)
}

// hookPreCompact handles the PreCompact hook.
// The compactor command is passed as os.Args[3:] (after "hook pre-compact").
func hookPreCompact() error {
	if len(os.Args) < 4 {
		return fmt.Errorf("no compactor command provided")
	}
	compactor := strings.Join(os.Args[3:], " ")

	input, err := readHookInput()
	if err != nil {
		return err
	}
	debugLog("pre-compact: session_id=%s transcript=%s", input.SessionID, input.TranscriptPath)

	if input.SessionID == "" {
		return fmt.Errorf("no session_id in hook input")
	}

	if input.TranscriptPath == "" {
		return fmt.Errorf("no transcript_path in hook input")
	}

	if _, err := os.Stat(input.TranscriptPath); err != nil {
		return fmt.Errorf("transcript not found: %s", input.TranscriptPath)
	}

	// Parse the transcript into plaintext for the compactor's stdin
	formatted, err := formatTranscript(input.TranscriptPath)
	if err != nil {
		return fmt.Errorf("failed to format transcript: %w", err)
	}
	debugLog("pre-compact: transcript_bytes=%d", len(formatted))

	// Run the compactor with plaintext on stdin and JSONL path as $1
	cmdStr := compactor + " " + shellQuote(input.TranscriptPath)
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdin = strings.NewReader(formatted)
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	output, err := cmd.Output()
	if stderrBuf.Len() > 0 {
		debugLog("pre-compact: compactor stderr: %s", stderrBuf.String())
	}
	if err != nil {
		return fmt.Errorf("compactor failed: %w", err)
	}
	debugLog("pre-compact: compactor output_bytes=%d", len(output))

	// Save output for session-start to pick up
	sf, err := stateFile(input.SessionID)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(sf), 0o755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	if err := os.WriteFile(sf, output, 0o644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// hookSessionStart handles the SessionStart hook (compact matcher).
func hookSessionStart() error {
	input, err := readHookInput()
	if err != nil {
		return err
	}

	if input.SessionID == "" {
		return fmt.Errorf("no session_id in hook input")
	}

	sf, err := stateFile(input.SessionID)
	if err != nil {
		return err
	}
	debugLog("session-start: session_id=%s state_file=%s", input.SessionID, sf)

	data, err := os.ReadFile(sf)
	if err != nil {
		if os.IsNotExist(err) {
			debugLog("session-start: no state file found")
			return nil
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}
	debugLog("session-start: injecting %d bytes", len(data))

	// Return as JSON with additionalContext field
	output := map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":   "SessionStart",
			"additionalContext": string(data),
		},
	}
	json.NewEncoder(os.Stdout).Encode(output)

	// Clean up
	os.Remove(sf)

	return nil
}
