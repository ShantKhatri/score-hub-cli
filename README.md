# score-hub

**The CLI install layer for the Score community provisioners ecosystem.**

Replace 140-character raw GitHub URLs with a single install command.

```bash
# Before score-hub:
score-k8s init --provisioners \
  https://raw.githubusercontent.com/score-spec/community-provisioners/refs/heads/main/dapr-pubsub/score-k8s/10-redis-dapr-pubsub.provisioners.yaml

# After score-hub:
score-hub install dapr-pubsub --variant redis
```

score-hub is to [community-provisioners](https://github.com/score-spec/community-provisioners) what [Krew](https://krew.sigs.k8s.io/) is to kubectl plugins.

## Quick Start (60 seconds)

```bash
# 1. Install score-hub
go install github.com/ShantKhatri/score-hub-cli@latest

# 2. Search for provisioners
score-hub search

# 3. Get details on a provisioner
score-hub info dapr-pubsub

# 4. Install into your Score project
cd my-score-project
score-hub install dapr-pubsub --variant redis

# 5. Use normally with score-k8s
score-k8s generate score.yaml
```

## Commands

### `score-hub search [query]`

Search for provisioners by name, category, or keyword. If no query is provided, all available provisioners are grouped by category.

```bash
score-hub search               # list all provisioners grouped by category
score-hub search dapr          # search by name
score-hub search messaging     # search by category
score-hub search --platform k8s  # filter by platform
score-hub search --json        # JSON output
```

### `score-hub info <name>`

Show detailed information about a provisioner including variants, prerequisites, and install commands.

```bash
score-hub info dapr-pubsub
score-hub info route --json
score-hub info llm-model
```

### `score-hub install <name> [flags]`

Install a provisioner into your Score project directory.

```bash
score-hub install dapr-pubsub --variant redis
score-hub install dapr-pubsub --variant redis --platform k8s
score-hub install dapr-pubsub@1.0.0 --variant redis
score-hub install environment --yes    # skip confirmation
```

Flags:
- `--variant` — Select a specific variant (required if multiple exist)
- `--platform` — Override platform detection (k8s or compose)
- `--yes` — Skip confirmation prompts
- `--no-verify` — Skip checksum verification (not recommended)
- `--dir` — Override install directory

### `score-hub list`

List installed provisioners in the current project.

```bash
score-hub list
score-hub list --json
```

## Platform Detection

score-hub automatically detects your Score platform in this priority:

1. `--platform` flag
2. `SCORE_HUB_PLATFORM` environment variable
3. `~/.score-hub/config.yaml` platform setting
4. `.score-k8s/` directory in current directory
5. `.score-compose/` directory in current directory
6. `score-k8s` binary in PATH
7. `score-compose` binary in PATH

## Lock File

score-hub creates a `.score-hub.lock` file tracking installed provisioners with version pins and checksums. **Commit this to source control** for reproducible provisioner installs.

## Configuration

Optional config at `~/.score-hub/config.yaml`:

```yaml
apiVersion: score-hub/v1alpha1
registry: https://raw.githubusercontent.com/score-hub/index/main/index.yaml
platform: auto  # k8s | compose | auto
cache:
  ttl: 1h
  dir: ~/.score-hub/cache/
```

## Global Flags

```
--platform string    Score platform: k8s or compose (default: auto-detect)
--json               Output in JSON format
--no-cache           Skip local index cache, fetch fresh
--verbose            Verbose output
--registry string    Override registry URL
```

## Available Provisioners

| Name | Category | Platforms | Variants |
|------|----------|-----------|----------|
| dapr-pubsub | messaging | k8s, compose | rabbitmq, redis |
| dapr-state-store | messaging | k8s, compose | redis |
| dapr-subscription | messaging | k8s, compose | default |
| dns | networking | k8s, compose | codespace, url |
| environment | configuration | k8s, compose | default |
| gcp-pubsub-emulator | messaging | compose | default |
| horizontal-pod-autoscaler | compute | k8s, compose | default |
| llm-model | llm | compose | dmr, dmr-openai, dmr-curl-cmd, dmr-curl-service, dmr-service-provider, ollama |
| redis | storage | k8s | helm-template, helm-upgrade |
| route | networking | k8s | ingress, ingress-netpol, gateway-httproute, gateway-httproute-netpol |
| service | networking | k8s, compose | default, microcks, microcks-cli, netpol |

## Architecture

score-hub uses a **Resolver** interface abstraction that decouples CLI commands from the data backend. v0.1 ships with a GitHub-backed resolver. The architecture supports future backends (local filesystem, HTTP API, private registries).

```
score-hub CLI → Resolver Interface → GitHub (community-provisioners)
                                   → Local (v0.2)
                                   → HTTP API (v0.3)
```

## Building from Source

```bash
git clone https://github.com/ShantKhatri/score-hub-cli.git
cd cli
make build
```

## Index Maintenance

When provisioners are updated in the upstream repository, the checksums in `score-hub` must be synced to ensure verification passes. We provide a utility script to do this automatically:

```bash
# Print a diff report of any upstream changes
go run scripts/sync_checksums.go

# Auto-update the index.yaml file in place
go run scripts/sync_checksums.go --update
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute.

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.

## Related

- [Score Specification](https://score.dev)
- [community-provisioners](https://github.com/score-spec/community-provisioners)
- [score-k8s](https://github.com/score-spec/score-k8s)
- [score-compose](https://github.com/score-spec/score-compose)
