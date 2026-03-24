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
)

const usage = `Usage: compact-hook <command> [args]

Commands:
  install --compactor <cmd>   Install hooks and compact instructions
  uninstall                   Remove hooks and compact instructions
  hook                        Called by Claude Code (not for direct use)

Flags:
  --global         Install/uninstall in ~/.claude/ instead of locally

Options:
  --help, -h       Show this help
  --version, -v    Show version
`

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		compactor := ""
		global := false
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--compactor":
				if i+1 < len(os.Args) {
					compactor = os.Args[i+1]
					i++
				}
			case "--global":
				global = true
			}
		}
		if compactor == "" {
			fmt.Fprintln(os.Stderr, "compact-hook: --compactor <command> is required")
			os.Exit(1)
		}
		if err := cmdInstall(compactor, global); err != nil {
			fmt.Fprintf(os.Stderr, "compact-hook: %v\n", err)
			os.Exit(1)
		}
	case "uninstall":
		global := false
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--global" {
				global = true
			}
		}
		if err := cmdUninstall(global); err != nil {
			fmt.Fprintf(os.Stderr, "compact-hook: %v\n", err)
			os.Exit(1)
		}
	case "hook":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "compact-hook: hook requires a subcommand (pre-compact, session-start)")
			os.Exit(1)
		}
		// Hook commands never exit non-zero to avoid blocking Claude Code
		switch os.Args[2] {
		case "pre-compact":
			if err := hookPreCompact(); err != nil {
				fmt.Fprintf(os.Stderr, "compact-hook: pre-compact: %v\n", err)
			}
		case "session-start":
			if err := hookSessionStart(); err != nil {
				fmt.Fprintf(os.Stderr, "compact-hook: session-start: %v\n", err)
			}
		default:
			fmt.Fprintf(os.Stderr, "compact-hook: unknown hook subcommand: %s\n", os.Args[2])
		}
	case "--help", "-h":
		fmt.Print(usage)
	case "--version", "-v":
		fmt.Println(version)
	default:
		fmt.Fprintf(os.Stderr, "compact-hook: unknown command: %s\n", os.Args[1])
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
}
