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
	"path/filepath"
)

// cmdInstall installs hooks and compact instructions.
func cmdInstall(compactor string, global bool) error {
	claudeDir, err := resolveClaudeDir(global)
	if err != nil {
		return err
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")

	if err := installHooks(settingsPath, compactor); err != nil {
		return fmt.Errorf("failed to install hooks: %w", err)
	}
	fmt.Printf("Installed hooks in %s\n", settingsPath)

	mdPath, err := installCompactInstructions(claudeDir)
	if err != nil {
		return fmt.Errorf("failed to install compact instructions: %w", err)
	}
	fmt.Printf("Added compact instructions to %s\n", mdPath)

	return nil
}

// cmdUninstall removes hooks and compact instructions.
func cmdUninstall(global bool) error {
	claudeDir, err := resolveClaudeDir(global)
	if err != nil {
		return err
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")

	if err := uninstallHooks(settingsPath); err != nil {
		return fmt.Errorf("failed to uninstall hooks: %w", err)
	}
	fmt.Printf("Removed hooks from %s\n", settingsPath)

	if err := uninstallCompactInstructions(claudeDir); err != nil {
		return fmt.Errorf("failed to remove compact instructions: %w", err)
	}
	fmt.Println("Removed compact instructions from CLAUDE.md")

	return nil
}
