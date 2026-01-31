# xpand Architecture Overview

## Vision

xpand implements the "Configuration as Data" philosophy: configuration is explicit data that multiple tools can read and write—not hidden behind abstractions.

**Problem xpand solves:**
- Developers need infrastructure but shouldn't need to know AWS/Kubernetes details
- Traditional abstractions (Helm, Compositions) hide too much—creating the "200% knowledge problem"
- Configuration should be a shared substrate, not owned by one team

**xpand's approach: Progressive disclosure, not abstraction.**
- Start with minimal input (2-5 fields)
- Generate fully explicit resources
- Developers CAN see everything, but don't HAVE to
- Any tool can modify any field (security, finops, SRE)

## System Context

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Developer Workflow                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                        ┌──────────────┼──────────────┐
                        │              │              │
                        ▼              ▼              ▼
                   ┌────────┐    ┌─────────┐    ┌────────────┐
                   │  CLI   │    │  Spec   │    │ Interactive│
                   │ Flags  │    │  File   │    │   Prompt   │
                   └────┬───┘    └────┬────┘    └─────┬──────┘
                        │             │               │
                        └─────────────┼───────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                                  xpand                                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   Pattern    │  │    Input     │  │  Expansion   │  │   Output     │    │
│  │   Loader     │──│  Validator   │──│   Engine     │──│  Formatter   │    │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘    │
│         │                                    │                              │
│         ▼                                    ▼                              │
│  ┌──────────────┐                    ┌──────────────┐                      │
│  │   Pattern    │                    │    Merge     │                      │
│  │   Catalog    │                    │    Engine    │                      │
│  └──────────────┘                    └──────────────┘                      │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
                        ┌─────────────────────────┐
                        │    Explicit YAML Files  │
                        │  (storage.yaml, iam.yaml│
                        │   compute.yaml, etc.)   │
                        └────────────┬────────────┘
                                     │
                                     ▼
                              ┌─────────────┐
                              │     Git     │
                              └──────┬──────┘
                                     │
                                     ▼
                              ┌─────────────┐
                              │ CI Pipeline │
                              └──────┬──────┘
                                     │
                    ┌────────────────┼────────────────┐
                    ▼                ▼                ▼
             ┌──────────┐     ┌──────────┐     ┌──────────┐
             │ConfigHub │     │  ArgoCD  │     │  Flux    │
             └──────────┘     └──────────┘     └──────────┘
                    │                │                │
                    └────────────────┼────────────────┘
                                     ▼
                              ┌─────────────┐
                              │  Crossplane │
                              │  (Actuator) │
                              └──────┬──────┘
                                     │
                                     ▼
                              ┌─────────────┐
                              │     AWS     │
                              └─────────────┘
```

## Core Components

### Pattern Loader

Loads pattern definitions from the pattern source (embedded, files, or registry—see [ADR-001](decisions/001-pattern-storage.md)).

Responsibilities:
- Discover available patterns
- Parse pattern definitions
- Validate pattern schema
- Cache patterns for performance

### Input Validator

Validates developer inputs against pattern schema.

Responsibilities:
- Type checking (string, int, bool)
- Required field validation
- Enum validation
- Pattern matching (regex)
- Custom validation rules
- Clear error messages

### Expansion Engine

Core logic that transforms pattern + inputs into explicit resources.

Responsibilities:
- Apply inputs to pattern template
- Generate all resources
- Set field ownership markers
- Ensure deterministic output
- Handle environment variants (see [ADR-003](decisions/003-environment-variants.md))

### Merge Engine

Handles updates when resources already exist.

Responsibilities:
- Compare existing vs. generated
- Preserve external changes
- Detect conflicts
- Apply user overrides (--set)

See [ADR-005](decisions/005-merge-preserve.md) for details.

### Output Formatter

Formats expanded resources for output.

Responsibilities:
- Group resources by type
- Generate multi-doc YAML
- Add metadata comments
- Support multiple formats (yaml, json)

## Data Flow

### Create Flow (New Resources)

```
1. User: xpand create serverless-event-app --env dev --account 123...
2. CLI parses flags into Inputs struct
3. Pattern Loader fetches pattern definition
4. Input Validator checks inputs against pattern schema
5. Expansion Engine generates resources from pattern + inputs
6. Output Formatter writes grouped YAML files
7. User commits to Git
8. CI pipeline validates and pushes to target
```

### Update Flow (Existing Resources)

```
1. User: xpand create serverless-event-app --env dev --account 123...
2. CLI parses flags into Inputs struct
3. Pattern Loader fetches pattern definition
4. Input Validator checks inputs
5. xpand reads existing resources from disk
6. Expansion Engine generates fresh resources
7. Merge Engine compares fresh vs. existing
8. Merge Engine preserves external changes, reports conflicts
9. Output Formatter writes merged resources
10. User commits to Git
```

## Pattern Structure

A pattern definition contains:

```
┌─────────────────────────────────────────────────┐
│                    Pattern                       │
├─────────────────────────────────────────────────┤
│ Metadata                                         │
│   - name: serverless-event-app                  │
│   - version: 1.0.0                              │
│   - description: ...                            │
├─────────────────────────────────────────────────┤
│ Input Schema                                     │
│   - env: string (required, enum)                │
│   - account: string (required, pattern)         │
│   - memory: int (optional, default: 128)        │
│   - ...                                         │
├─────────────────────────────────────────────────┤
│ Resource Definitions                             │
│   - Bucket (S3)                                 │
│   - Table (DynamoDB)                            │
│   - Role × 2 (IAM)                              │
│   - Policy × 2 (IAM)                            │
│   - Function × 2 (Lambda)                       │
│   - FunctionURL (Lambda)                        │
│   - Rule + Target (EventBridge)                 │
│   - ...                                         │
├─────────────────────────────────────────────────┤
│ Field Ownership                                  │
│   - Which fields are owned by xpand            │
│   - Which fields are safe for external changes │
└─────────────────────────────────────────────────┘
```

See [ADR-002](decisions/002-pattern-format.md) for format details.

## Output Structure

xpand generates files grouped by resource type:

```
output/
├── storage.yaml      # S3 Bucket, DynamoDB Table
├── iam.yaml          # IAM Roles, Policies
├── compute.yaml      # Lambda Functions, FunctionURL
├── events.yaml       # EventBridge Rule, Target, Permission
└── .xpand-meta.yaml  # Generation metadata (optional)
```

Each file is multi-document YAML:

```yaml
# storage.yaml
# Generated by xpand v1.0.0
# Pattern: serverless-event-app
# Timestamp: 2026-01-31T10:00:00Z
---
apiVersion: s3.aws.upbound.io/v1beta2
kind: Bucket
metadata:
  name: myapp-dev-123456789012-bucket
  labels:
    xpand.io/pattern: serverless-event-app
    xpand.io/version: "1.0.0"
spec:
  forProvider:
    region: us-east-1
---
apiVersion: dynamodb.aws.upbound.io/v1beta2
kind: Table
# ...
```

## Key Design Decisions

| Decision | Status | Document |
|----------|--------|----------|
| Pattern storage | Proposed | [ADR-001](decisions/001-pattern-storage.md) |
| Pattern format | Proposed | [ADR-002](decisions/002-pattern-format.md) |
| Environment variants | Proposed | [ADR-003](decisions/003-environment-variants.md) |
| AI agent as user | Proposed | [ADR-004](decisions/004-ai-agent-user.md) |
| Merge/preserve | Proposed | [ADR-005](decisions/005-merge-preserve.md) |

## Future Considerations

- **Pattern versioning**: How to handle breaking changes in patterns
- **Pattern composition**: Composing multiple patterns together
- **Import from existing**: Generate pattern from existing resources
- **Dry-run against live**: Compare generated vs. ConfigHub/cluster state
- **IDE integration**: VS Code extension for pattern authoring
