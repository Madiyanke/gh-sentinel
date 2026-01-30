# 🛡️ Sentinel CI - The DevOps Guardian

> **Stop guessing why your build failed.** Diagnose, repair, and revive CI/CD pipelines directly from your terminal.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![GitHub CLI](https://img.shields.io/badge/GitHub_CLI-Extension-181717?logo=github)](https://cli.github.com)

**Built for the [GitHub Copilot CLI Challenge 2026](https://dev.to/challenges/github-copilot-cli)**

## Why I Built Sentinel CI

Most CI tools just tell you *that* you failed. Sentinel CI tells you *why* and asks: **"Do you want me to fix it?"**

Unlike simple linters, Sentinel CI is an autonomous agent designed to solve the "Red Cross" nightmare:

* **Auto-Discovery**: Instantly finds failed workflows in your current repo context.
* **Noise Reduction**: Focuses strictly on your latest commit failures.
* **Deep Diagnostics**: Uses the GitHub Copilot engine to understand root causes, not just error codes.
* **Precision Targeting**: Corrects the *actual* broken file, even if the error logs point elsewhere.
* **Surgical Patching**: Rewrites valid YAML configurations locally—no broken snippets.
* **Safety First**: Creates automatic backups (`.bak`) before touching a single line of code.
* **Developer UX**: A clean, interactive TUI built with Bubble Tea that respects your terminal workflow.
* **Universal**: Runs natively on Windows, Linux, and macOS without complex setup.

## See it in Action

```bash
$ gh sentinel
```

## Quick Start

### Prerequisites

1. **GitHub CLI** installed and authenticated:
```bash
gh auth login
```

2. **GitHub Copilot Extension**:
```bash
gh extension install github/gh-copilot
```

### Installation

#### Recommended: Install as `gh` extension

```bash
gh extension install Madiyanke/gh-sentinel
```

#### Alternative: Build from source

```bash
git clone https://github.com/Madiyanke/gh-sentinel.git
cd gh-sentinel
go build -o gh-sentinel.exe ./cmd/sentinel
gh extension install .
```

### Usage

Navigate to any local git repository with GitHub Actions failures and run:

```bash
gh sentinel
```

Sentinel CI will:

1. Detect your repository context.
2. Scan for recent failures.
3. Analyze logs using pattern matching & Copilot intelligence.
4. Propose a precise fix with a diff view.
5. Apply the patch locally upon your confirmation.

## Architecture

I designed Sentinel CI with an **industrial-grade modular architecture** to ensure stability and maintainability:

```text
gh-sentinel/
├── cmd/sentinel/          # Entry point
├── internal/
│   ├── config/           # Configuration logic
│   ├── context/          # Universal repo & auth detection
│   ├── errors/           # Custom typed error handling
│   ├── logger/           # Structured logging system
│   ├── orchestrator/     # Core workflow logic
│   └── ui/               # Bubble Tea TUI components
└── pkg/
    ├── analyzer/         # RegEx-based log pre-analysis
    ├── copilot/          # Programmatic Copilot integration
    ├── github/           # GitHub API wrapper
    └── patcher/          # Safe file I/O with backups
```

### Engineering Decisions

#### 1. Zero Hard Dependencies

I used `gh repo view` and `gh auth token` strategies to ensure the tool works instantly in any user's environment. No config files, no hardcoded paths.

#### 2. Robust Error Handling

No `panic()`. Every error is typed, contextualized, and presented clearly to the user. Path handling uses `filepath` for full cross-platform compatibility.

#### 3. Multi-Stage Intelligence

* **Stage 1 (Fast)**: Regex pattern matching for common errors (Node versions, missing secrets, etc.).
* **Stage 2 (Deep)**: Copilot analysis with engineered system prompts for logic errors.
* **Stage 3 (Verify)**: User diff review before application.

#### 4. Safety Mechanisms

* Automatic timestamped backups (`.sentinel.bak`).
* YAML validation.
* Rollback capability.

## Advanced Capabilities

### Intelligent Target Redirection

Sentinel CI is smart enough to know when the error isn't where you think it is.

```text
Target Redirection Detected:
  User Selected: .github/workflows/deploy.yml
  Actual Culprit: .github/workflows/build.yml (Based on log evidence)
```

### Forensic Pre-Analysis

Before consulting Copilot, the tool runs a quick forensic scan:

```text
Detected Potential Issues:
  1. Deprecation Warning: Node.js 12 actions are deprecated
  2. Exit Code 1: Process completed with error
  3. Module Error: npm ERR! code ENOENT
```

## Development

### Building & Testing

```bash
# Standard build
go build -o gh-sentinel.exe ./cmd/sentinel

# Production build (stripped)
go build -ldflags="-s -w" -o gh-sentinel.exe ./cmd/sentinel

# Run tests
go test ./...
```

### Code Quality

The project adheres to strict Go standards:

* **Linting**: `golangci-lint`
* **Formatting**: `gofmt`
* **Vet**: `go vet`

## Project Stats

* **Lines of Code**: ~2,500
* **Modules**: 11
* **Error Patterns**: 10+
* **Platform Support**: Windows, Linux, macOS

## Contributing

Contributions are welcome! Please ensure:

1. Code compiles without errors.
2. No panics are introduced.
3. New features include tests.
4. Go idioms are respected.

See [CONTRIBUTING.md](./CONTRIBUTING.md).

## License

MIT License - see [LICENSE](./LICENSE).

## Acknowledgments

* **GitHub Copilot Team** for the powerful CLI engine.
* **Charm Bracelet** for the Bubble Tea framework.
* **GitHub CLI Team** for the extensible `gh` platform.

## Submission Context

This project demonstrates:

* **Deep Copilot Integration**: Moving beyond chat to agentic behavior.
* **Exceptional UX**: A TUI that feels native and responsive.
* **Production Quality**: Engineered for reliability and safety.
* **Real Value**: Solves a daily pain point for every DevOps engineer.

---

<div align="center">

**Built with ❤️ and Go by [Madiyanke](https://github.com/Madiyanke)**

*Making CI/CD failures a thing of the past.*

[Report Bug](https://github.com/Madiyanke/gh-sentinel/issues) • [Request Feature](https://github.com/Madiyanke/gh-sentinel/issues)

</div>