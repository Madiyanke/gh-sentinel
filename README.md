# 🛡️ Sentinel-CI
> Your terminal's new sidekick for GitHub Actions health.

Sentinel-CI is a `gh` extension that uses GitHub Copilot CLI to diagnose and repair your CI/CD pipelines directly from your terminal.

## Features
- **Zero-Friction Discovery**: Automatically finds recent failed runs in your current repo.
- **AI-Powered Diagnostics**: Get a clear explanation of what went wrong.
- **One-Click Repair**: Apply the suggested YAML patch locally.
- **Safety First**: Automatic `.bak` file creation before any modification.

## Installation
```bash
git clone [https://github.com/Madiyanke/gh-sentinel](https://github.com/Madiyanke/gh-sentinel)
cd gh-sentinel
go build -o gh-sentinel.exe ./cmd/sentinel
gh extension install .