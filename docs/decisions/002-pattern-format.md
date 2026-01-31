# ADR-002: Pattern Definition Format

**Status:** Proposed
**Date:** 2026-01-31

## Context

Platform teams need to define patterns that xpand expands into explicit resources. We need to decide the format for these pattern definitions.

A pattern definition must specify:
- **Metadata**: Name, description, version, author
- **Input schema**: What fields developers provide (name, type, default, validation)
- **Resource templates**: How inputs map to Kubernetes resources
- **Field ownership**: Which fields xpand owns vs. can be modified by other tools
- **Documentation**: Help text, examples

## Decision Drivers

- **Authoring experience**: How easy is it for platform teams to create patterns?
- **Expressiveness**: Can it handle complex resource generation?
- **Validation**: Can inputs be validated before expansion?
- **Tooling**: Are there editors, linters, formatters available?
- **Learning curve**: How much must pattern authors learn?
- **Testability**: Can patterns be unit tested?
- **AI agent authoring**: Could AI agents write patterns?

## Considered Options

### Option A: Go Code (Compiled)

Patterns are Go packages that programmatically generate resources.

```go
type ServerlessEventAppPattern struct{}

func (p *ServerlessEventAppPattern) InputSchema() InputSchema {
    return InputSchema{
        Fields: []Field{
            {Name: "env", Type: "string", Required: true, Enum: []string{"dev", "staging", "prod"}},
            {Name: "accountId", Type: "string", Required: true, Pattern: "^[0-9]{12}$"},
            // ...
        },
    }
}

func (p *ServerlessEventAppPattern) Expand(inputs map[string]any) ([]Resource, error) {
    bucket := s3.Bucket{
        Metadata: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s-bucket", inputs["env"], inputs["accountId"])},
        Spec: s3.BucketSpec{ForProvider: s3.BucketParameters{Region: inputs["region"].(string)}},
    }
    // ... generate all resources
    return []Resource{bucket, table, ...}, nil
}
```

**Pros:**
- Full programming language power
- Type safety and IDE support
- Easy testing with Go test framework
- Can use Crossplane Go types directly
- Complex logic is straightforward

**Cons:**
- Requires Go knowledge
- Patterns compiled into binary (unless plugins)
- Higher barrier for platform teams
- Harder for AI to generate

### Option B: YAML Templates

Patterns are YAML files with simple templating (like Helm but simpler).

```yaml
name: serverless-event-app
version: 1.0.0
description: Serverless event-driven application

inputs:
  - name: env
    type: string
    required: true
    enum: [dev, staging, prod]
  - name: accountId
    type: string
    required: true
    pattern: "^[0-9]{12}$"

resources:
  - apiVersion: s3.aws.upbound.io/v1beta2
    kind: Bucket
    metadata:
      name: "{{ .env }}-{{ .accountId }}-bucket"
    spec:
      forProvider:
        region: "{{ .region | default \"us-east-1\" }}"
```

**Pros:**
- Familiar to Helm users
- Easy to read and modify
- No compilation needed
- Lower barrier for platform teams
- AI can easily generate/modify

**Cons:**
- Limited expressiveness (conditionals, loops)
- Template debugging is hard
- "Templates in YAML" is awkward
- Creates mini-language to learn

### Option C: CUE Language

Use CUE for schema definition and resource generation.

```cue
package serverlesseventapp

#Input: {
    env:       "dev" | "staging" | "prod"
    accountId: =~"^[0-9]{12}$"
    region:    *"us-east-1" | "us-west-2" | "eu-west-1"
    memory:    int & >=128 & <=10240 | *128
}

#Bucket: {
    input: #Input
    apiVersion: "s3.aws.upbound.io/v1beta2"
    kind: "Bucket"
    metadata: name: "\(input.env)-\(input.accountId)-bucket"
    spec: forProvider: region: input.region
}

resources: [#Bucket & {input: _input}, #Table & {input: _input}, ...]
```

**Pros:**
- Designed for configuration
- Strong typing and validation built-in
- No separate schema vs template
- Composable and importable
- Growing ecosystem

**Cons:**
- New language to learn
- Less familiar than YAML/Go
- Tooling still maturing
- Complex logic can be awkward

### Option D: KCL (Kusion Configuration Language)

Use KCL, designed specifically for Kubernetes configuration.

**Pros:**
- Kubernetes-native design
- Python-like syntax (familiar)
- Built-in Kubernetes schema support
- IDE support available

**Cons:**
- Smaller community than CUE
- Learning new language
- Less portable outside Kubernetes

### Option E: Pkl (Apple's Configuration Language)

Use Pkl, Apple's new configuration language.

**Pros:**
- Modern design
- Strong typing
- IDE support (JetBrains, VS Code)
- Active development

**Cons:**
- Very new (2024)
- Smaller ecosystem
- Unknown long-term trajectory

## Decision

*[To be decided]*

## Consequences

### If Go Code
- Platform teams need Go skills
- Maximum flexibility but higher barrier
- Patterns tightly coupled to xpand releases

### If YAML Templates
- Low barrier but limited expressiveness
- Familiar to most teams
- Template debugging challenges

### If CUE/KCL/Pkl
- Modern approach with strong typing
- Learning curve for new language
- Good long-term investment if ecosystem grows

## Notes

Consider:
- The serverless-event-app pattern has 17 resources with complex IAM policies—can the chosen format handle this elegantly?
- AI agents should ideally be able to both USE patterns and CREATE patterns
- Pattern testing should be straightforward
- Start with one pattern, evolve format as we learn
