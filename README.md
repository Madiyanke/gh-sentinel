# 🛡️ Sentinel CI - AI-Powered DevOps Guardian

> **Your terminal's ultimate sidekick for GitHub Actions** - Diagnose, repair, and revive failed CI/CD pipelines with the power of AI.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev/)
[![GitHub CLI](https://img.shields.io/badge/GitHub_CLI-Required-181717?logo=github)](https://cli.github.com)

**Built for the [GitHub Copilot CLI Challenge 2026](https://dev.to/challenges/github-copilot-cli)** 🏆

## ✨ What Makes Sentinel CI Different?

Unlike simple linters or log viewers, **Sentinel CI is an autonomous DevOps agent** that:

- 🔍 **Auto-discovers** failed workflows in your current repository
- 🎯 **Latest commit focus** - Shows only failures from your most recent push (ignores historical noise)
- 🧠 **AI-powered diagnosis** using GitHub Copilot for root cause analysis  
- 🎯 **Target redirection** - Identifies the *actual* problematic file, not just the symptom
- 🔧 **Surgical fixes** - Applies complete, working YAML corrections (no placeholders!)
- 💾 **Safety-first** - Automatic backups before every modification
- 🎨 **Beautiful TUI** - Interactive terminal interface built with Bubble Tea
- 🌍 **Universal** - Works on Windows, Linux, macOS without any hardcoded paths

## 🎬 Demo

```bash
$ gh sentinel
```

![Demo](./assets/demo.gif)

## 🚀 Quick Start

### Prerequisites

1. **GitHub CLI** installed and authenticated
   ```bash
   # Install gh CLI: https://cli.github.com
   gh auth login
   ```

2. **GitHub Copilot CLI extension**
   ```bash
   gh extension install github/gh-copilot
   ```

### Installation

#### Option 1: Install as gh extension (Recommended)
```bash
gh extension install Madiyanke/gh-sentinel
```

#### Option 2: Build from source
```bash
git clone https://github.com/Madiyanke/gh-sentinel.git
cd gh-sentinel
go build -o gh-sentinel.exe ./cmd/sentinel
gh extension install .
```

### Usage

Simply navigate to any Git repository and run:

```bash
gh sentinel
```

That's it! Sentinel CI will:
1. Detect your repository automatically
2. Scan for failed workflow runs
3. Let you select which failure to investigate
4. Analyze logs with pattern matching
5. Consult GitHub Copilot for AI diagnosis
6. Present a fix with diff preview
7. Apply the patch after your confirmation

## 🏗️ Architecture

Sentinel CI follows an **industrial-grade modular architecture**:

```
gh-sentinel/
├── cmd/sentinel/          # Application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── context/          # Repository & auth detection
│   ├── errors/           # Typed error handling
│   ├── logger/           # Structured logging
│   ├── orchestrator/     # Main workflow coordinator
│   └── ui/               # Terminal UI components (Bubble Tea)
└── pkg/
    ├── analyzer/         # Pattern-based log analysis
    ├── copilot/          # GitHub Copilot integration
    ├── github/           # GitHub API client
    └── patcher/          # Safe file patching with backups
```

### Key Design Decisions

#### 1. **Zero Hard-Coded Dependencies**
- Uses `gh repo view` and `gh auth token` for universal portability
- Works in ANY repository without configuration
- No hardcoded GitHub usernames or paths

#### 2. **Robust Error Handling**
- Custom typed errors with full context
- No `panic()` calls - graceful degradation everywhere
- Cross-platform path handling via `path/filepath`

#### 3. **Multi-Stage Intelligence**
- **Stage 1**: Pattern matching (10+ common CI/CD error patterns)
- **Stage 2**: AI diagnosis via GitHub Copilot with engineered prompts
- **Stage 3**: Validation & diff preview before applying changes

#### 4. **Safety Mechanisms**
- Automatic timestamped backups (`.sentinel.bak`)
- YAML validation before writing
- Rollback capability
- Diff preview for user verification

## 🎯 Advanced Features

### Intelligent Target Detection

Sentinel CI doesn't just trust the failing workflow file—it examines logs to find the **actual culprit**:

```
⚠ Target Redirection Detected
  User Selected: .github/workflows/deploy.yml
  AI Identified: .github/workflows/build.yml (Based on log evidence)
```

### Pattern-Based Pre-Analysis

Before calling Copilot, Sentinel runs pattern matching to provide instant insights:

```
🔍 Detected 3 potential issues:
  1. Node.js Version Deprecated: Node.js 12 actions are deprecated
  2. Exit Code Non-Zero: Process completed with exit code 1
  3. NPM Install Failed: npm ERR! code ENOENT

💡 Quick Suggestions:
  1. Update to a newer Node.js version in your workflow
  2. Check the command output above for the actual error
  3. Check package.json or run npm install locally first
```

### Comprehensive Logging

```
[05:52:12] INFO: Authenticated as repository: Madiyanke/test-sentinel-ci
[05:52:13] DEBUG: Retrieved 5 workflow runs
[05:52:14] INFO: Found 2 failed runs out of 5 total
[05:52:16] DEBUG: Retrieved logs from 1 failed jobs
```

## 🛠️ Development

### Building

```bash
# Development build
go build -o gh-sentinel.exe ./cmd/sentinel

# Production build (optimized)
go build -ldflags="-s -w" -o gh-sentinel.exe ./cmd/sentinel

# Cross-compilation examples
GOOS=linux GOARCH=amd64 go build -o gh-sentinel-linux ./cmd/sentinel
GOOS=darwin GOARCH=arm64 go build -o gh-sentinel-mac ./cmd/sentinel
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/analyzer -v
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint (requires golangci-lint)
golangci-lint run

# Vet code
go vet ./...
```

## 📊 Project Stats

- **Lines of Code**: ~2,500 (excluding comments/blanks)
- **Modules**: 11 production modules
- **Error Patterns**: 10+ CI/CD-specific patterns
- **Dependencies**: Minimal (Bubble Tea, go-github, oauth2)
- **Platform Support**: Windows, Linux, macOS

## 🤝 Contributing

We welcome contributions! This project follows strict quality standards:

1. **Code must compile** with zero errors
2. **Error handling** - No panics, always return errors
3. **Documentation** - All exported functions must have comments
4. **Testing** - New features require tests
5. **Style** - Follow `gofmt` and idiomatic Go practices

See [CONTRIBUTING.md](./CONTRIBUTING.md) for detailed guidelines.

## 📄 License

MIT License - see [LICENSE](./LICENSE) for details.

## 🙏 Acknowledgments

- **GitHub Copilot Team** - For the amazing AI coding assistant
- **Charm Bracelet** - For the beautiful Bubble Tea TUI framework
- **GitHub CLI Team** - For the powerful `gh` command-line tool

## 🏆 GitHub Copilot CLI Challenge 2026

This project was built for the **GitHub Copilot CLI Challenge**. It demonstrates:

✅ **Innovative Copilot CLI Integration** - Goes beyond simple queries to create an autonomous agent  
✅ **Exceptional UX** - Beautiful, interactive TUI with real-time feedback  
✅ **Production Quality** - Industrial-grade error handling, logging, and safety mechanisms  
✅ **Real-World Impact** - Solves actual DevOps pain points developers face daily  

---

<div align="center">

**Built with ❤️ and lots of ☕ by [Madiyanke](https://github.com/Madiyanke)**

*Making CI/CD failures a thing of the past, one deployment at a time* 🚀

[Report Bug](https://github.com/Madiyanke/gh-sentinel/issues) • [Request Feature](https://github.com/Madiyanke/gh-sentinel/issues) • [Documentation](https://github.com/Madiyanke/gh-sentinel/wiki)

</div>