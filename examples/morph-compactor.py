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

"""Example compactor using Morph's compact API.

Setup:
    compact-hook install --compactor "python3 /path/to/morph-compactor.py"

Receives formatted plaintext transcript on stdin.
Uses Morph's line-deletion compaction (no rewriting) at 33,000 tok/s.

Requires: pip install requests
Environment: MORPH_API_KEY
"""

import os
import sys

import requests

MORPH_API_KEY = os.environ.get("MORPH_API_KEY", "")
MORPH_API_URL = "https://api.morphllm.com/v1/compact"


def main():
    if not MORPH_API_KEY:
        print("morph-compactor: MORPH_API_KEY not set", file=sys.stderr)
        sys.exit(1)

    transcript = sys.stdin.read()
    if not transcript.strip():
        print("Empty session - no prior context.")
        return

    response = requests.post(
        MORPH_API_URL,
        headers={"Authorization": f"Bearer {MORPH_API_KEY}"},
        json={
            "input": transcript,
            "compression_ratio": 0.5,
            "preserve_recent": 3,
            "include_markers": False,
        },
    )
    response.raise_for_status()

    data = response.json()
    print(data["output"])


if __name__ == "__main__":
    main()
