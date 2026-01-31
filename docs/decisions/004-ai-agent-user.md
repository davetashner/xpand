# ADR-004: AI Agent as Primary User

**Status:** Proposed
**Date:** 2026-01-31

## Context

xpand will be used equally by humans and AI agents (Claude Code, GitHub Copilot, etc.). This is not an afterthought—AI agents are a primary user persona from day one.

This affects:
- CLI design (flags, help text, output formats)
- Error messages
- Discoverability
- Determinism
- Documentation

## Decision Drivers

- **Agent discoverability**: Can an agent learn to use xpand from --help alone?
- **Structured output**: Can agents parse xpand output reliably?
- **Determinism**: Same inputs always produce same outputs?
- **Error clarity**: Do error messages help agents self-correct?
- **Human usability**: Does agent-friendliness hurt human experience?

## Design Principles

Based on how AI agents interact with CLIs, we propose these design principles:

### 1. Comprehensive Help Text

Every command and flag must have:
- Clear description of what it does
- Description of expected input format
- At least one example showing common usage
- Description of output format

**Good:**
```
--account string
    AWS account ID (12 digits). Used for resource naming and IAM ARNs.
    Example: --account 123456789012
```

**Bad:**
```
--account string    AWS account
```

### 2. Structured Output

All commands should support `--output json` for machine-readable output.

**Good:**
```json
{
  "status": "success",
  "files": [
    {"path": "storage.yaml", "resources": 2},
    {"path": "iam.yaml", "resources": 4}
  ],
  "totalResources": 17
}
```

**Bad:**
```
Created 17 resources in 4 files.
```

### 3. Deterministic Behavior

Same inputs must always produce same outputs:
- No random IDs unless explicitly requested
- No timestamps in generated content (or make them optional)
- Stable ordering of resources and fields
- Predictable file names

### 4. Actionable Error Messages

Errors should explain:
- What went wrong
- Why it's a problem
- How to fix it

**Good:**
```
Error: Invalid account ID "12345"
  Account ID must be exactly 12 digits (got 5 digits).
  Example: --account 123456789012
```

**Bad:**
```
Error: invalid input
```

### 5. Schema Discovery

Agents should be able to discover:
- What patterns are available
- What inputs each pattern requires
- What defaults exist
- What validation rules apply

```bash
xpand patterns describe serverless-event-app --output json
```

```json
{
  "name": "serverless-event-app",
  "inputs": [
    {
      "name": "env",
      "type": "string",
      "required": true,
      "enum": ["dev", "staging", "prod"],
      "description": "Deployment environment"
    },
    {
      "name": "account",
      "type": "string",
      "required": true,
      "pattern": "^[0-9]{12}$",
      "description": "AWS account ID (12 digits)"
    }
  ]
}
```

### 6. Progressive Disclosure

Start simple, allow complexity:
- Minimal required inputs (2-3 fields)
- Sensible defaults for everything else
- Advanced flags for customization
- `--dry-run` to preview without side effects

### 7. Exit Codes

Meaningful exit codes for automation:
- 0: Success
- 1: General error
- 2: Invalid input
- 3: Conflict detected (merge needed)
- 4: Pattern not found

## Considered Options

### Option A: Agent-First Design

Optimize primarily for agent usage; humans adapt.

**Pros:**
- Cleanest agent experience
- Forces good structure
- Future-proof as agents improve

**Cons:**
- May feel verbose to humans
- Unusual CLI conventions

### Option B: Human-First with Agent Support

Traditional CLI design with `--output json` bolted on.

**Pros:**
- Familiar to humans
- Lower risk

**Cons:**
- Agents treated as second-class
- Inconsistent structured output

### Option C: Dual-Mode (Recommended)

Design for both equally. Human-friendly defaults with agent-friendly options.

**Pros:**
- Best of both worlds
- No compromise for either persona

**Cons:**
- More design effort
- Must test both paths

## Decision

We recommend **Option C: Dual-Mode**.

- Default output is human-readable text
- `--output json` provides structured output
- Help text is comprehensive (benefits humans too)
- Examples in help (benefits humans too)
- Actionable errors (benefits humans too)

Most agent-friendly features also benefit humans.

## Consequences

### Positive
- Agents can use xpand effectively
- Humans get better documentation and errors
- Future-proof for increasing AI assistance

### Negative
- More effort to maintain dual output paths
- Help text is longer (but more useful)
- Testing matrix includes human + agent scenarios

### Neutral
- Sets a precedent for other tools in the ecosystem

## Implementation Notes

### Cobra Best Practices

```go
var createCmd = &cobra.Command{
    Use:   "create <pattern> [flags]",
    Short: "Create Kubernetes resources from a pattern",
    Long: `Create explicit Kubernetes resources by expanding a pattern with your inputs.

xpand reads the pattern definition, validates your inputs, and generates
fully-explicit Kubernetes YAML files grouped by resource type.

The generated files are suitable for committing to Git and applying via
GitOps tools (ArgoCD, Flux) or directly with kubectl.

Patterns:
  Use 'xpand patterns list' to see available patterns.
  Use 'xpand patterns describe <name>' for pattern details.

Examples:
  # Create serverless app with minimal input
  xpand create serverless-event-app --env dev --account 123456789012

  # Create with custom settings
  xpand create serverless-event-app --env prod --account 123456789012 \
    --set lambda.memory=512 --set lambda.timeout=30

  # Preview without writing files
  xpand create serverless-event-app --env dev --account 123456789012 --dry-run

  # Output as JSON for scripting
  xpand create serverless-event-app --env dev --account 123456789012 --output json`,
    Example: `  xpand create serverless-event-app --env dev --account 123456789012`,
    Args: cobra.ExactArgs(1),
    RunE: runCreate,
}
```

### Testing for Agents

Include tests that simulate agent usage:

```go
func TestCreateCommand_AgentUsage(t *testing.T) {
    // Agent discovers pattern schema
    schemaOut := runCommand(t, "xpand", "patterns", "describe", "serverless-event-app", "--output", "json")
    var schema PatternSchema
    json.Unmarshal(schemaOut, &schema)

    // Agent builds command from schema
    args := []string{"xpand", "create", schema.Name}
    for _, input := range schema.Inputs {
        if input.Required {
            args = append(args, fmt.Sprintf("--%s", input.Name), testValueFor(input))
        }
    }
    args = append(args, "--output", "json")

    // Agent runs command
    result := runCommand(t, args...)

    // Agent parses result
    var createResult CreateResult
    require.NoError(t, json.Unmarshal(result, &createResult))
    assert.Equal(t, "success", createResult.Status)
}
```
