# xpand

**Expand application intent into explicit Kubernetes resources.**

xpand is a CLI tool that takes minimal developer input and generates fully-explicit Kubernetes manifests (typically Crossplane ManagedResources). It implements the "Configuration as Data" philosophy: no runtime rendering, no hidden abstractions—what you see is what gets deployed.

## Status

**Pre-alpha** - Design phase. See [docs/decisions/](docs/decisions/) for ADRs and [.beads/backlog.jsonl](.beads/backlog.jsonl) for the development backlog.

## Philosophy

Traditional infrastructure tools create a "200% knowledge problem": developers must understand both the output format (Kubernetes resources) AND the abstraction layer that generates them (Helm charts, Terraform modules, Crossplane Compositions).

xpand takes a different approach: **progressive disclosure, not abstraction**.

- Developers start with minimal input (2-5 fields)
- xpand generates explicit, fully-rendered resources
- Developers CAN see everything, but don't HAVE to
- Any tool (security scanners, cost optimizers, AI agents) can modify any field

## How It Works

```
Developer → xpand (progressive disclosure) → Explicit Resources → Git → CI Pipeline → Target
                                                                              ↓
                                                              (ConfigHub, ArgoCD, etc.)
```

### Quick Example

```bash
# Generate serverless event-driven app resources
xpand create serverless-event-app \
  --env dev \
  --account 205074708100 \
  --region us-east-1

# Output: Grouped YAML files
# - storage.yaml (S3, DynamoDB)
# - iam.yaml (Roles, Policies)
# - compute.yaml (Lambda functions)
# - events.yaml (EventBridge)
```

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Output flow | Git-first | Resources go to Git, CI pipeline pushes to target |
| Update semantics | Merge/preserve | Don't overwrite fields modified by other tools |
| Primary users | Humans AND AI agents | Equally usable by both |
| Pattern catalog | General with patterns | Platform teams define patterns |
| XRD relationship | Optional alternative | Organizations can choose xpand OR Composition |

## Documentation

- [Architecture Overview](docs/architecture.md)
- [Decision Records](docs/decisions/)
- [Pattern Development Guide](docs/patterns.md) *(planned)*
- [CLI Reference](docs/cli-reference.md) *(planned)*

## Development

```bash
# Build
go build -o xpand ./cmd/xpand

# Test
go test ./...

# Run
./xpand --help
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

Apache 2.0 - See [LICENSE](LICENSE)
