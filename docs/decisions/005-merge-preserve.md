# ADR-005: xpand Scope - Initial Generation Only

**Status:** Proposed
**Date:** 2026-01-31
**Supersedes:** Original merge/preserve semantics discussion

## Context

Originally, we planned for xpand to handle both initial generation AND updates with complex merge/preserve semantics. After reviewing:

1. Brian Grant's ["Configuration needs an API"](https://itnext.io/configuration-needs-an-api-b36f08b92551)
2. The [ConfigHub SDK function model](https://github.com/confighub/sdk/blob/main/function/README.md)

We recognize a cleaner separation of concerns that aligns with the Configuration as Data philosophy.

## The Key Insight

> "Functions and other tools built around the API can be single purpose or multi-purpose, but they do not need to be monolithic in the way that configuration generators do."
> — Brian Grant

**Configuration has an API.** Once resources exist, modifications should use API-based tools, not re-generation.

## Decision Drivers

- **Separation of concerns**: Generation is different from mutation
- **Composability**: ConfigHub functions compose; generators don't
- **Reduced complexity**: No merge logic in xpand
- **Ecosystem alignment**: Use ConfigHub's existing function infrastructure
- **Multi-source support**: Functions enable Security/FinOps/SRE to modify without understanding xpand

## Considered Options

### Option A: xpand Handles Everything (Original Design)

xpand handles both initial creation and updates with merge/preserve logic.

```bash
# Initial
xpand create serverless-event-app --env dev ...

# Update (xpand must merge)
xpand create serverless-event-app --env dev --set timeout=30
```

**Pros:**
- Single tool for everything
- Familiar "re-run to update" pattern

**Cons:**
- Complex merge/preserve logic
- Must track field ownership
- Duplicates ConfigHub function capabilities
- Monolithic, not composable

### Option B: xpand for Initial, Functions for Updates (Recommended)

xpand only generates initial resources. All subsequent changes use ConfigHub functions.

```bash
# Initial generation (xpand)
xpand create serverless-event-app --env dev --account 123...

# Subsequent changes (ConfigHub functions)
cub fn invoke set-lambda-memory --unit api-handler --value 256
cub fn invoke set-env-var --name LOG_LEVEL --value debug
cub fn invoke set-tag --key cost-center --value platform
```

**Pros:**
- Clean separation of concerns
- Leverages ConfigHub's composable function model
- No merge complexity in xpand
- Functions work for any tool (Security, FinOps, SRE)
- Aligns with "configuration as API" philosophy

**Cons:**
- Two tools to learn (xpand + ConfigHub functions)
- Initial setup requires both

### Option C: xpand Generates Function Sequences

xpand outputs a sequence of ConfigHub function invocations instead of raw YAML.

```bash
xpand create serverless-event-app --env dev --account 123... --output functions
```

Output:
```yaml
functions:
  - name: create-bucket
    args: {name: "messagewall-dev-123...", region: "us-east-1"}
  - name: create-table
    args: {name: "messagewall-dev-123...", region: "us-east-1"}
  # ... etc
```

**Pros:**
- Everything is functions
- Unified model

**Cons:**
- Requires "create-X" functions for every resource type
- More complex than just generating YAML
- Initial creation doesn't benefit from composability

## Decision

**Option B: xpand for initial generation, ConfigHub functions for all updates.**

### xpand's Scope

| In Scope | Out of Scope |
|----------|--------------|
| Generate resources from scratch | Update existing resources |
| Apply pattern defaults | Merge with existing values |
| Validate inputs | Track field ownership |
| Output explicit YAML | Handle multi-source conflicts |

### Update Workflow

After initial `xpand create`:

```bash
# Developer changes timeout
cub fn invoke set-lambda-timeout --selector "kind=Function" --value 30

# Security adds compliance tag
cub fn invoke set-tag --key security-reviewed --value "2026-01-31"

# FinOps right-sizes memory
cub fn invoke set-lambda-memory --unit api-handler --value 192

# SRE adds debug env var during incident
cub fn invoke set-env-var --name DEBUG_MODE --value true
```

All changes:
- Use the configuration API (paths, selectors)
- Are composable with other functions
- Don't require understanding xpand patterns
- Are tracked in ConfigHub revision history

### What About Re-Running xpand?

If a developer runs `xpand create` again on existing resources:

1. **Default behavior**: Error with message directing to ConfigHub functions
2. **`--force` flag**: Regenerate everything (destructive, warns loudly)
3. **`--diff` flag**: Show what would be different without applying

```bash
$ xpand create serverless-event-app --env dev ...
Error: Resources already exist in output directory.

To modify existing resources, use ConfigHub functions:
  cub fn invoke set-lambda-memory --unit api-handler --value 256
  cub fn invoke set-env-var --name KEY --value value

To see what would change: xpand create ... --diff
To force regeneration (DESTRUCTIVE): xpand create ... --force
```

## Consequences

### Positive
- xpand becomes much simpler (no merge logic)
- Clear mental model: xpand = bootstrap, functions = modify
- Leverages ConfigHub's mature function infrastructure
- Multi-source modifications work naturally
- Field ownership tracking not needed in xpand

### Negative
- Users must learn ConfigHub functions for updates
- Two tools instead of one
- Initial documentation must explain the split

### Neutral
- Aligns with broader ConfigHub ecosystem direction
- May influence how other generation tools integrate

## Implementation Notes

### Remove from xpand backlog
- EPIC-4 (Merge/Preserve System) - no longer needed
- ISSUE-4.1, 4.2, 4.3 - delete

### Add to xpand
- Clear error when resources exist
- `--diff` flag to compare
- `--force` flag (with confirmation) for regeneration
- Documentation pointing to ConfigHub functions

### Consider for ConfigHub
- Ensure functions exist for common Crossplane resource modifications
- `set-lambda-memory`, `set-lambda-timeout` for Lambda Functions
- `set-dynamodb-capacity` for DynamoDB Tables
- Generic `set-string-path`, `set-int-path` as fallbacks
