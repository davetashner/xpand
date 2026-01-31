# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

xpand is a Go CLI tool that generates explicit Kubernetes resources from minimal developer input. It implements the "Configuration as Data" (CaD) philosophy—no runtime rendering, no hidden abstractions.

**Key principle:** Progressive disclosure, not abstraction. Developers start simple, can dig deeper when needed, and any tool can modify the explicit output.

## Architecture

```
Patterns (YAML/Go)     Developer Input        Output
       ↓                     ↓                  ↓
  ┌─────────┐          ┌──────────┐      ┌─────────────┐
  │ Pattern │ ──────── │  xpand   │ ──── │ Explicit    │ ─── Git ─── CI ─── Target
  │ Catalog │          │   CLI    │      │ K8s YAML    │
  └─────────┘          └──────────┘      └─────────────┘
                             ↑
                       Spec file or
                       CLI flags or
                       Interactive
```

## Development Commands

```bash
# Build
go build -o xpand ./cmd/xpand

# Test
go test ./...

# Test with coverage
go test -cover ./...

# Lint
golangci-lint run

# Run locally
go run ./cmd/xpand --help
```

## Code Organization

```
xpand/
├── cmd/xpand/          # CLI entry point (cobra commands)
├── internal/
│   ├── pattern/        # Pattern loading and expansion
│   ├── output/         # YAML output formatting
│   ├── merge/          # Merge/preserve logic for updates
│   └── spec/           # Spec file parsing
├── patterns/           # Built-in pattern definitions
├── docs/
│   └── decisions/      # ADRs (Architecture Decision Records)
└── .beads/             # Issue tracking backlog
```

## Commit Message Format

Use Conventional Commits:

```
<type>(<scope>): <subject>

Types: feat, fix, docs, refactor, test, chore
Scope: cli, pattern, output, merge, spec, adr, backlog
```

Examples:
```
feat(cli): Add create command with pattern selection
fix(merge): Preserve fields with mw.xpand.io/preserve annotation
docs(adr): Add ADR-001 for pattern storage decision
```

## Backlog Management

This project uses [Beads](https://github.com/steveyegge/beads) for issue tracking.

```bash
bd list              # View all issues
bd show <issue-id>   # View issue details
bd create "title"    # Create new issue
bd update <id> --status in_progress
```

## ADR (Architecture Decision Record) Format

ADRs live in `docs/decisions/` and follow this format:

```markdown
# ADR-NNN: Title

**Status:** Proposed | Accepted | Deprecated | Superseded
**Date:** YYYY-MM-DD

## Context
What is the issue that we're seeing that is motivating this decision?

## Decision Drivers
- Driver 1
- Driver 2

## Considered Options
1. Option A
2. Option B
3. Option C

## Decision
What is the change that we're proposing and/or doing?

## Consequences
### Positive
- ...

### Negative
- ...

### Neutral
- ...
```

## Agent-Friendly Design Principles

xpand is designed to be equally usable by humans and AI agents. When developing:

1. **Clear help text**: Every command and flag should have descriptive help that an agent can understand
2. **Structured output**: Support `--output json` for machine-readable output
3. **Predictable behavior**: Same input always produces same output (deterministic)
4. **Good error messages**: Errors should explain what went wrong AND suggest fixes
5. **Examples in help**: Include usage examples in command help text

Example of good help text:
```go
cmd := &cobra.Command{
    Use:   "create <pattern> [flags]",
    Short: "Create resources from a pattern",
    Long: `Create Kubernetes resources by expanding a pattern with your inputs.

The pattern name must be one of the available patterns (use 'xpand patterns list').
Output is written to the current directory grouped by resource type.

Examples:
  # Create serverless app with minimal input
  xpand create serverless-event-app --env dev --account 123456789012

  # Create with custom memory and timeout
  xpand create serverless-event-app --env prod --account 123456789012 \
    --set lambda.memory=512 --set lambda.timeout=30

  # Preview without writing files
  xpand create serverless-event-app --env dev --account 123456789012 --dry-run`,
}
```

## Testing Requirements

All code should have tests:

1. **Unit tests**: For individual functions
2. **Integration tests**: For command execution
3. **Golden file tests**: For YAML output (expected output in testdata/)
4. **Pattern tests**: Each pattern should have test cases

## Key Design Decisions (Summary)

See `docs/decisions/` for full ADRs.

| # | Decision | Status |
|---|----------|--------|
| 001 | Pattern storage and distribution | Proposed |
| 002 | Pattern definition format | Proposed |
| 003 | Environment variant handling | Proposed |
| 004 | AI agent as primary user | Proposed |
