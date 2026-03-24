# Copyright 2026 StayBlue
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Example compactor using OpenRouter with Grok 4.1 Fast.

Setup:
    compact-hook install --compactor "python3 /path/to/openrouter-compactor.py"

Receives formatted plaintext transcript on stdin.
The raw JSONL path is also available as $1 if needed.

Requires: pip install requests
Environment: OPENROUTER_API_KEY
"""

import os
import sys

import requests

OPENROUTER_API_KEY = os.environ.get("OPENROUTER_API_KEY", "")
OPENROUTER_API_URL = "https://openrouter.ai/api/v1/chat/completions"
MODEL = "x-ai/grok-4.1-fast"
MAX_TOKENS = 8192
SYSTEM_PROMPT = """You are a conversation compactor. You will receive a Claude Code session transcript.
Produce a concise summary that preserves:
- What the user asked for and why
- Key decisions made and their rationale
- Current state of the work (what's done, what's pending)
- Important file paths, function names, and code context
- Any errors encountered and how they were resolved

Be concise but complete. The summary will replace the full conversation history."""


def main():
    if not OPENROUTER_API_KEY:
        print("openrouter-compactor: OPENROUTER_API_KEY not set", file=sys.stderr)
        sys.exit(1)

    transcript = sys.stdin.read()
    if not transcript.strip():
        print("Empty session - no prior context.")
        return

    response = requests.post(
        OPENROUTER_API_URL,
        headers={
            "Authorization": f"Bearer {OPENROUTER_API_KEY}",
            "Content-Type": "application/json",
        },
        json={
            "model": MODEL,
            "messages": [
                {"role": "system", "content": SYSTEM_PROMPT},
                {"role": "user", "content": transcript},
            ],
            "max_tokens": MAX_TOKENS,
            "temperature": 0.2,
        },
        timeout=120,
    )
    response.raise_for_status()

    data = response.json()
    content = data["choices"][0]["message"]["content"]

    if isinstance(content, list):
        text_parts = [
            part.get("text", "")
            for part in content
            if isinstance(part, dict) and part.get("type") == "text"
        ]
        print("".join(text_parts))
        return

    print(content)


if __name__ == "__main__":
    main()
