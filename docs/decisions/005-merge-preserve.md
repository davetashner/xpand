# ADR-005: Merge/Preserve Semantics for Updates

**Status:** Proposed
**Date:** 2026-01-31

## Context

When xpand runs against resources that already exist, it must decide what to do. The key constraint: **other tools (security scanners, cost optimizers, SRE) may have modified fields, and those changes should not be lost.**

This is central to the Configuration as Data philosophy—configuration is a shared substrate that multiple tools write to.

Example scenario:
1. Developer runs `xpand create` → generates Lambda with memory=128
2. FinOps runs cost optimizer → changes memory to 192 (right-sized)
3. Developer runs `xpand create` again (maybe changing timeout)
4. **Question**: What happens to memory?
   - If xpand overwrites: memory reverts to 128 (FinOps change lost)
   - If xpand preserves: memory stays 192 (correct)

## Decision Drivers

- **Multi-source safety**: Don't overwrite changes from other tools
- **Predictability**: Users should understand what xpand will change
- **Conflict detection**: Alert when xpand and another tool disagree
- **Auditability**: Track who owns which fields
- **Simplicity**: Don't make the update flow overly complex

## Field Ownership Model

We propose a field ownership model where each field has an "owner":

- **xpand-owned**: Fields that xpand sets from pattern logic (e.g., region derived from --region flag)
- **external-owned**: Fields set by other tools (security tags, cost optimizations)
- **user-owned**: Fields explicitly set by user via --set flag

### Ownership Tracking Options

#### Option A: Annotations

Track ownership in Kubernetes annotations:

```yaml
metadata:
  annotations:
    xpand.io/field-owners: |
      spec.forProvider.region: xpand
      spec.forProvider.memorySize: finops-optimizer
      spec.forProvider.tags.security-reviewed: security-scanner
```

**Pros:**
- Visible in the resource
- Standard Kubernetes pattern
- Tools can read ownership

**Cons:**
- Annotation can get large
- Must parse on every operation
- Synchronization challenges

#### Option B: Sidecar File

Track ownership in a separate file alongside resources:

```yaml
# .xpand-ownership.yaml
resources:
  - name: api-handler
    kind: Function
    fields:
      spec.forProvider.region: xpand
      spec.forProvider.memorySize: external
      spec.forProvider.tags.security-reviewed: external
```

**Pros:**
- Doesn't clutter resources
- Easy to parse
- Can track history

**Cons:**
- Extra file to manage
- Can get out of sync
- Not visible in ConfigHub/GitOps

#### Option C: Diff-Based Detection

Don't track ownership explicitly. Instead, compare current state to what xpand would generate:

1. Read existing resources
2. Generate what xpand would produce
3. Diff to find fields that differ
4. Preserve differing fields (assume external ownership)

**Pros:**
- No ownership tracking needed
- Works with existing resources
- Simple mental model

**Cons:**
- Can't distinguish "user changed" from "external changed"
- False positives if pattern defaults change
- Less precise conflict detection

## Merge Strategies

### Strategy 1: Preserve External (Recommended)

xpand only updates fields it "owns". Fields that differ from xpand's output are assumed external and preserved.

```
Existing:   memorySize: 192  (set by finops)
xpand wants: memorySize: 128  (pattern default)
Result:     memorySize: 192  (preserved)
```

### Strategy 2: Warn on Conflict

Same as Strategy 1, but warn when preservation happens:

```
Warning: Preserving external changes to api-handler:
  - spec.forProvider.memorySize: 192 (xpand default: 128)
  Use --overwrite to use xpand defaults instead.
```

### Strategy 3: Explicit Ownership Flags

User explicitly declares what to preserve:

```bash
xpand create ... --preserve spec.forProvider.memorySize --preserve spec.forProvider.tags
```

**Pros:**
- Maximum control
- No magic

**Cons:**
- Verbose
- Easy to forget fields

## Conflict Scenarios

### Scenario A: External Modified, User Wants Different

```
Existing:    memorySize: 192 (finops)
User wants:  memorySize: 256 (via --set)
```

**Resolution**: User's explicit --set takes precedence.

### Scenario B: Pattern Changed, External Modified

Pattern v1 default: 128MB
Pattern v2 default: 256MB
External set: 192MB

**Resolution**: Preserve external (192MB). User must explicitly --set to override.

### Scenario C: Conflicting External Tools

Security set: tag.security-level=high
Cost optimizer: tag.security-level=medium

**Resolution**: This is outside xpand's scope. Last writer wins in Git. xpand preserves whatever exists.

## Decision

We recommend:

1. **Option C (Diff-Based Detection)** for ownership tracking—simplest implementation, works with existing resources
2. **Strategy 2 (Warn on Conflict)** for merge behavior—preserve external changes but inform user
3. **User override via --set** takes precedence over both xpand defaults and external changes
4. **--overwrite flag** to forcibly regenerate everything (dangerous, requires confirmation)

## Consequences

### Positive
- External changes are preserved by default
- Users are informed when preservation happens
- Works with existing resources without migration
- --set provides explicit override

### Negative
- May preserve unintentional changes
- Pattern default changes don't propagate automatically
- Warning messages may be noisy initially

### Neutral
- Users must use --set or --overwrite for intentional changes
- Responsibility for field values is distributed

## Implementation Notes

### Diff-Based Merge Algorithm

```go
func (e *Expander) MergeWithExisting(pattern string, inputs Inputs, existing Resource) (Resource, []Conflict) {
    // Generate what xpand would produce from scratch
    fresh := e.Expand(pattern, inputs)

    // Start with existing as base
    result := existing.DeepCopy()

    var conflicts []Conflict

    // For each field xpand wants to set
    for path, freshValue := range fresh.FlattenFields() {
        existingValue := existing.GetField(path)

        // If field was explicitly set via --set, always use new value
        if inputs.ExplicitlySet(path) {
            result.SetField(path, inputs.Get(path))
            continue
        }

        // If existing matches fresh, no conflict
        if existingValue == freshValue {
            continue
        }

        // Existing differs from fresh—preserve existing, record conflict
        conflicts = append(conflicts, Conflict{
            Path:          path,
            ExistingValue: existingValue,
            XpandValue:    freshValue,
        })
        // result already has existingValue, so nothing to do
    }

    return result, conflicts
}
```

### Warning Output

```
$ xpand create serverless-event-app --env dev --account 123...

Preserving external changes (use --overwrite to discard):
  api-handler:
    - spec.forProvider.memorySize: 192 (xpand default: 128)
    - spec.forProvider.tags.cost-center: platform (not in pattern)
  snapshot-writer:
    - spec.forProvider.timeout: 45 (xpand default: 10)

Created 17 resources in 4 files.
```
