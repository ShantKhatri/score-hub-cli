# Contributing to score-hub

Thank you for your interest in contributing to score-hub!

## Ways to Contribute

### Update the Index

The most impactful contribution is keeping the index up to date with community-provisioners.

1. Check [community-provisioners](https://github.com/score-spec/community-provisioners) for new provisioners
2. Add an entry to `data/index.yaml` following the existing format
3. Copy the updated file to `cmd/index.yaml`
4. Submit a PR

### Index Entry Format

```yaml
- name: my-provisioner
  displayName: "Human Readable Name"
  description: >
    A 2-3 sentence description of what this provisioner does.
  category: messaging|networking|storage|compute|llm|configuration
  tags: [relevant, tags]
  upstream: score-spec/community-provisioners
  variants:
    - id: variant-name
      displayName: "Variant Display Name"
      description: "What makes this variant different"
      platforms:
        k8s:
          path: directory/score-k8s/filename.provisioners.yaml
          filename: filename.provisioners.yaml
          checksum: ""
      prerequisites:
        - "Any prerequisites needed"
  versions:
    - version: "1.0.0"
      date: "2026-01-01"
      upstreamCommit: "commit-sha"
      changelog: "What changed"
```

### Report Bugs

Open an issue with:
- Steps to reproduce
- Expected vs actual behavior
- score-hub version (`score-hub version`)
- OS and architecture

### Improve Code

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Submit a PR

## Code Style

- Follow standard Go conventions
- Run `gofmt` before committing
- All CLI commands go through the Resolver interface (never call GitHub directly)

## Review SLA

We aim to review all PRs within 7 days.

## License

By contributing, you agree that your contributions will be licensed under Apache License 2.0.
