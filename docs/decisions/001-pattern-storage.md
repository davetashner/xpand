# ADR-001: Pattern Storage and Distribution

**Status:** Proposed
**Date:** 2026-01-31

## Context

xpand uses patterns to define how developer intent maps to explicit Kubernetes resources. Platform teams create these patterns. We need to decide how patterns are stored, versioned, and distributed to xpand users.

This decision affects:
- How platform teams author and publish patterns
- How xpand loads patterns at runtime
- How organizations distribute internal patterns
- How patterns are versioned and updated
- Offline usage capabilities

## Decision Drivers

- **Platform team authoring experience**: How easy is it to create and test patterns?
- **Distribution flexibility**: Can organizations distribute internal patterns?
- **Versioning**: Can users pin to specific pattern versions?
- **Offline usage**: Does xpand work without network access?
- **Update mechanism**: How do users get pattern updates?
- **Security**: Can organizations control which patterns are available?

## Considered Options

### Option A: Embedded in Binary

Patterns are compiled into the xpand Go binary.

**Pros:**
- Simplest distribution (single binary)
- Always available offline
- Version is implicit (pattern version = xpand version)
- No network dependencies

**Cons:**
- New patterns require xpand release
- Organizations cannot add internal patterns without forking
- Large binary if many patterns
- Pattern authors must know Go (if patterns are Go code)

### Option B: Local Pattern Files

Patterns are YAML/JSON files on the local filesystem. xpand reads from a patterns directory.

**Pros:**
- Easy to author and test locally
- Organizations can add internal patterns
- No network dependencies
- Clear separation of tool and patterns

**Cons:**
- Distribution is manual (copy files)
- Version management is user responsibility
- Patterns might get out of sync across team
- Need to define search paths

### Option C: Remote Registry

Patterns are fetched from a registry (ConfigHub, OCI registry, Git repo) at runtime.

**Pros:**
- Centralized pattern management
- Easy updates (pull latest)
- Organizations can host internal registries
- Versioning is explicit

**Cons:**
- Network dependency
- Latency on first use
- Need caching strategy for offline
- Registry availability becomes critical

### Option D: Hybrid (Embedded + Local + Remote)

Core patterns embedded, local overrides, optional remote registry.

**Pros:**
- Works offline with core patterns
- Organizations can extend
- Optional remote for enterprise features
- Flexibility for different use cases

**Cons:**
- Most complex implementation
- Precedence rules needed (which pattern wins?)
- More code paths to test
- User confusion about pattern sources

## Decision

*[To be decided]*

## Consequences

### If Option A (Embedded)
- Positive: Simplest user experience, always works
- Negative: Limited extensibility, release coupling

### If Option B (Local Files)
- Positive: Maximum flexibility, easy authoring
- Negative: Distribution burden on users

### If Option C (Remote Registry)
- Positive: Centralized management, easy updates
- Negative: Network dependency, operational overhead

### If Option D (Hybrid)
- Positive: Best of all worlds
- Negative: Implementation complexity

## Notes

Consider that AI agents will use xpand. Agents may:
- Run in ephemeral environments (fresh each time)
- Need deterministic behavior (same pattern version = same output)
- Not have access to internal registries

Also consider the relationship with ConfigHub:
- Could ConfigHub be the pattern registry?
- Does that create vendor coupling?
