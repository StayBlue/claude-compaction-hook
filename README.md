# claude-compaction-hook

This project was born out of a need to replace the garbage default compaction in Claude Code with something better.

> [!IMPORTANT]
> This is a massive hack. Until Anthropic provides a way to directly modify the compaction output via hooks, this is the best solution I could come up with.

The [Morph compaction model](https://www.morphllm.com/products/compact) seems to work well, and it's free if you make an account, but of course, you are free to use whatever model you like.

This project, along with the following documentation, is AI-generated. One thing to note that isn't noted below is that if you plan to compact early before auto compaction kicks in, you need to run the command with the custom instructions as well: `/compact Do not summarize the conversation. Output only the following and nothing else: Summary provided via SessionStart hook.`

## License

This project is licensed under the Apache License, Version 2.0. You are free to use this project as you see fit so long as you comply with the license's terms.

---

## What it does

When Claude Code compacts a session, this tool:

1. Adds a small `Compact Instructions` section to `CLAUDE.md` so Claude emits a placeholder instead of a full summary.
2. Runs your compactor in the `PreCompact` hook against the full transcript.
3. Injects your compactor's output back into context during `SessionStart`.

## Install

```bash
go install -trimpath -ldflags="-s -w" github.com/StayBlue/claude-compaction-hook@latest
```

Or build from source:

```bash
go build -trimpath -ldflags="-s -w" -o compact-hook .
```

## Use

Install in the current project:

```bash
compact-hook install --compactor "python3 /path/to/my-compactor.py"
```

Install globally:

```bash
compact-hook install --global --compactor "python3 /path/to/my-compactor.py"
```

`install` updates `.claude/settings.json` or `~/.claude/settings.json` and adds the required `Compact Instructions` section to `CLAUDE.md`.

Remove it with:

```bash
compact-hook uninstall
compact-hook uninstall --global
```

Run `install` again with a new `--compactor` value to replace the existing command.

## Writing a compactor

Your compactor can be any executable that:

- reads the formatted transcript from `stdin`
- receives the raw JSONL transcript path as `$1`
- writes the replacement summary to `stdout`
- exits with code `0` on success

The formatted transcript looks like this:

```text
[user]: ...
[assistant]: ...
```

See `examples/` for reference implementations.
