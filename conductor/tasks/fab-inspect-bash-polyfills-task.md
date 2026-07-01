# NATVS Task: Bash Shell Polyfills for IDE Agent Resilience

## Objective
Implement Go-native polyfills within `fab-inspect` (or a shared utility package in `s-fab-aides` named `pkg/bash`) that emulate common Bash commands. The IDE Agent (Antigravity) frequently defaults to executing standard Bash commands within Windows PowerShell environments, leading to immediate syntax failures and unnecessary recovery cycles. 

By providing sovereign Go implementations of these commands natively inside the codebase, we eliminate cross-platform shell incompatibilities entirely.

## Requirements

1. **Target Component:** `00flow/s-fab-aides/81000-active-source/cmd/fab-inspect`
2. **Implementations:**
   - Create robust Go functions in a `bash` package to emulate the behavior of:
     - `ls` (List directory contents, optionally with sizes/permissions)
     - `cat` (Print file contents to stdout)
     - `grep` (Search text patterns inside files using regex)
     - `cp` (Copy files or directories)
     - `mv` (Move/rename files or directories)
     - `rm` (Remove files or directories, support `-rf` behavior)
     - `touch` (Create empty files or update timestamps)
     - `mkdir` (Create directories, support `-p`)
     - `pwd` (Print current working directory)
     - `find` (Recursively search for files/directories matching patterns)
     - `sed` (Stream editor for text replacement)
     - `head` / `tail` (Output the first/last parts of files)
     - `echo` (Print text)
     - `tree` (List contents of directories in a tree-like format)
     - `chmod` / `chown` (Modify file permissions and ownership mappings)
     - `env` (List current environment variables)
     - `which` (Locate executables within the execution path)
     - `tar` / `zip` / `unzip` (Archive creation and extraction)
     - `df` / `du` (Report file system disk space usage and directory sizes)
     - `curl` / `wget` (Basic HTTP GET/POST operations for fetching remote artifacts)
3. **CLI Integration:**
   - Wire these functions into `fab-inspect` as a new top-level subcommand suite (e.g., `fab-inspect bash <cmd> [args]`) or as standalone commands if structurally appropriate.
4. **Resilience & Safety:**
   - Ensure the implementations handle Windows file paths gracefully.
   - Return clean exit codes so the IDE Agent can parse the results deterministically without shell-parsing errors.

## NATVS Execution Directives
- **Assimilation:** Ingest `main.go` inside `cmd/fab-inspect`.
- **Transformation:** Create a `pkg/bash` implementing the logic. Wire it into the `fab-inspect` CLI parser under a `bash` subcommand namespace.
- **Verification:** Ensure the commands execute correctly against local sandbox directories via automated `go test` coverage. **Critically, implement comprehensive unit test suites achieving a target of 100% code coverage for all polyfill functions in `pkg/bash`.**
- **Synthesis:** Seal the changes and promote to the internal `94000-internal-actors` registry via the `int-rehydrator.go` engine.
