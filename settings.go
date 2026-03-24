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
	"os"
	"path/filepath"
	"strings"
)

// resolveClaudeDir returns the .claude/ directory to install into.
// If global is true, uses ~/.claude/. Otherwise, uses .claude/ in the
// current directory, creating it if needed.
func resolveClaudeDir(global bool) (string, error) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		claudeDir := filepath.Join(home, ".claude")
		if _, err := os.Stat(claudeDir); err != nil {
			return "", fmt.Errorf("~/.claude/ does not exist")
		}
		return claudeDir, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	claudeDir := filepath.Join(cwd, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create %s: %w", claudeDir, err)
	}

	return claudeDir, nil
}

// hookCommand returns the absolute path to this binary followed by the given subcommand args.
func hookCommand(args ...string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot determine executable path: %w", err)
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("cannot resolve executable path: %w", err)
	}

	parts := append([]string{exe}, args...)
	return strings.Join(parts, " "), nil
}

// isOurHook checks if a hook entry contains our command.
func isOurHook(entry map[string]any) bool {
	hooks, ok := entry["hooks"].([]any)
	if !ok {
		return false
	}
	for _, h := range hooks {
		hm, ok := h.(map[string]any)
		if !ok {
			continue
		}
		cmd, _ := hm["command"].(string)
		if strings.Contains(cmd, "compact-hook hook") {
			return true
		}
	}
	return false
}

// installHooks adds our hook entries to settings.json.
func installHooks(settingsPath string, compactor string) error {
	preCompactCmd, err := hookCommand("hook", "pre-compact", compactor)
	if err != nil {
		return err
	}
	sessionStartCmd, err := hookCommand("hook", "session-start")
	if err != nil {
		return err
	}

	// Read existing settings or start fresh
	settings := make(map[string]any)
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse %s: %w", settingsPath, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", settingsPath, err)
	}

	// Get or create hooks map
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = make(map[string]any)
	}

	// Build our hook entries
	preCompactEntry := map[string]any{
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": preCompactCmd,
			},
		},
	}

	sessionStartEntry := map[string]any{
		"matcher": "compact",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": sessionStartCmd,
			},
		},
	}

	// Add PreCompact — replace existing or append
	preCompactList, _ := hooks["PreCompact"].([]any)
	preCompactList = replaceOrAppendHook(preCompactList, preCompactEntry)
	hooks["PreCompact"] = preCompactList

	// Add SessionStart — replace existing or append
	sessionStartList, _ := hooks["SessionStart"].([]any)
	sessionStartList = replaceOrAppendHook(sessionStartList, sessionStartEntry)
	hooks["SessionStart"] = sessionStartList

	settings["hooks"] = hooks

	return writeSettings(settingsPath, settings)
}

// uninstallHooks removes our hook entries from settings.json.
func uninstallHooks(settingsPath string) error {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read %s: %w", settingsPath, err)
	}

	settings := make(map[string]any)
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse %s: %w", settingsPath, err)
	}

	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		return nil
	}

	if list, ok := hooks["PreCompact"].([]any); ok {
		hooks["PreCompact"] = filterOutOurHooks(list)
		if len(hooks["PreCompact"].([]any)) == 0 {
			delete(hooks, "PreCompact")
		}
	}

	if list, ok := hooks["SessionStart"].([]any); ok {
		hooks["SessionStart"] = filterOutOurHooks(list)
		if len(hooks["SessionStart"].([]any)) == 0 {
			delete(hooks, "SessionStart")
		}
	}

	if len(hooks) == 0 {
		delete(settings, "hooks")
	}

	return writeSettings(settingsPath, settings)
}

// replaceOrAppendHook replaces an existing hook entry that's ours, or appends.
func replaceOrAppendHook(list []any, entry map[string]any) []any {
	for i, existing := range list {
		em, ok := existing.(map[string]any)
		if !ok {
			continue
		}
		if isOurHook(em) {
			list[i] = entry
			return list
		}
	}
	return append(list, entry)
}

// filterOutOurHooks returns a new list with our hook entries removed.
func filterOutOurHooks(list []any) []any {
	var result []any
	for _, entry := range list {
		em, ok := entry.(map[string]any)
		if !ok {
			result = append(result, entry)
			continue
		}
		if !isOurHook(em) {
			result = append(result, entry)
		}
	}
	if result == nil {
		result = []any{}
	}
	return result
}

// writeSettings writes the settings map back to the file with 2-space indent.
func writeSettings(path string, settings map[string]any) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	data = append(data, '\n')

	return os.WriteFile(path, data, 0o644)
}
