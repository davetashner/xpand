# ADR-003: Environment Variant Handling

**Status:** Proposed
**Date:** 2026-01-31

## Context

Applications need different configurations for different environments (dev, staging, prod). Examples:
- Dev: 128MB Lambda memory, 10s timeout, no encryption required
- Prod: 512MB Lambda memory, 60s timeout, encryption required, point-in-time recovery

We need to decide how xpand handles these environment variants.

## Decision Drivers

- **Separation of concerns**: Who owns environment differences—xpand, platform team, or app team?
- **Flexibility**: Can organizations customize what "prod" means?
- **Simplicity**: How many concepts must developers learn?
- **Existing tooling**: Does this integrate with Kustomize, Helm, or similar?
- **Auditability**: Can we see exactly what differs between environments?

## Considered Options

### Option A: xpand Handles Entirely

Pattern definitions include environment-specific defaults and logic.

```bash
# xpand produces different output based on --env
xpand create serverless-event-app --env dev --account 123...
xpand create serverless-event-app --env prod --account 123...
```

Pattern definition:
```yaml
inputs:
  - name: memory
    type: int
    default:
      dev: 128
      staging: 256
      prod: 512
  - name: encryption
    type: bool
    default:
      dev: false
      staging: true
      prod: true
```

**Pros:**
- Single tool, simple mental model
- Environment logic in one place (pattern)
- No additional tooling needed
- Easy to see what differs between envs

**Cons:**
- Pattern complexity increases
- Platform teams must anticipate all env differences
- Less flexibility for app-specific overrides
- Duplicates Kustomize functionality

### Option B: Kustomize Overlays (External)

xpand generates "base" resources. Kustomize overlays apply environment patches.

```
infra/
├── base/
│   └── (xpand output)
└── overlays/
    ├── dev/
    │   └── kustomization.yaml  # patches for dev
    └── prod/
        └── kustomization.yaml  # patches for prod
```

**Pros:**
- Separation of concerns (xpand = base, Kustomize = variants)
- Familiar to Kubernetes users
- Maximum flexibility
- Leverages existing tooling

**Cons:**
- Two tools to learn
- More files to manage
- Harder to see full picture
- Patches can be complex

### Option C: Both (xpand Defaults + Override Layer)

xpand has environment-aware defaults. Kustomize (or similar) can layer additional changes.

```bash
# xpand produces env-appropriate defaults
xpand create serverless-event-app --env prod --account 123...

# Output already has prod defaults (512MB, encryption, etc.)
# Kustomize can override further if needed
```

**Pros:**
- Best of both worlds
- Sensible defaults without extra work
- Flexibility for customization
- Clear responsibility: xpand = sensible start, overlays = customization

**Cons:**
- Must clearly document what xpand defaults vs. what to override
- Potential confusion about where to make changes
- Testing matrix increases

### Option D: Spec File Per Environment

Developers maintain separate spec files per environment.

```
specs/
├── dev.yaml      # env: dev, memory: 128, ...
├── staging.yaml  # env: staging, memory: 256, ...
└── prod.yaml     # env: prod, memory: 512, ...
```

```bash
xpand create serverless-event-app --spec specs/prod.yaml
```

**Pros:**
- Explicit—each env fully specified
- No magic defaults
- Easy to diff between environments
- WET philosophy (Write Every Time)

**Cons:**
- Duplication between spec files
- Changes must be made in multiple places
- More files to manage

## Decision

*[To be decided]*

## Consequences

### If xpand Handles Entirely (A)
- Simpler user experience
- Patterns become more complex
- Less flexibility for edge cases

### If Kustomize Overlays (B)
- Maximum flexibility
- Steeper learning curve
- Matches existing Kubernetes patterns

### If Both (C)
- Good defaults + flexibility
- Must document clearly
- Slightly higher complexity

### If Spec File Per Environment (D)
- Maximum explicitness
- More repetition
- Aligns with WET philosophy

## Notes

The WET (Write Every Time) philosophy suggests explicit configuration is better than implicit derivation. This might favor Option D (explicit spec files) or Option C (defaults + explicit overrides) over Option A (magic env-based defaults).

Consider:
- How do other tools handle this? (Terraform workspaces, Helm values files)
- What do target users (platform teams, developers) expect?
- How does this interact with GitOps workflows?
