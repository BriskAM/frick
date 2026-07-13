# Contributing to frick

Thank you for your interest in contributing to `frick`! This document outlines the guidelines and procedures for contributing to this project.

---

## Code of Conduct

By participating in this project, you agree to maintain a professional, respectful, and cooperative environment. Please ensure that all communications and code contributions are clear, polite, and free of unnecessary clutter (including emojis in code, comments, or commit messages).

---

## How to Contribute

### 1. Reporting Bugs & Requesting Features
- Open an issue on GitHub describing the bug or feature request.
- Provide clear steps to reproduce bugs, including your Shell type (Zsh/Bash), Operating System, and Go version.

### 2. Development Setup
To build and run `frick` locally:
1. Ensure you have Go 1.21 or later installed.
2. Clone the repository:
   ```bash
   git clone https://github.com/BriskAM/frick.git
   cd frick
   ```
3. Compile the binary:
   ```bash
   go build -o frick cmd/frick/main.go
   ```
4. Test the CLI manually:
   ```bash
   ./frick run --cmd "gti status" --exit 127
   ```

### 3. Coding Style & Formatting
- **Standard Formatting**: Run `go fmt ./...` before committing any code.
- **Linting**: Ensure your code has no unused imports or variables.
- **Documentation**: Write clear, descriptive comments. Do not use emoji characters in code files, comments, documentation, or commit messages.

### 4. Committing Changes
- Write concise, professional commit messages.
- Avoid using emojis in commit titles or descriptions.
- Example of a good commit message:
  `Add local safety overrides feature`

### 5. Submitting Pull Requests
1. Fork the repository and create your branch from `main`.
2. Ensure your changes compile successfully.
3. Submit a Pull Request with a clear description of the problem solved or the feature added.

---

## Project Architecture

- `cmd/frick/main.go`: The main CLI entry point. Implements command parsing, raw TTY terminal manipulation, interactive arrow-key selector, and safety level double-confirmations.
- `internal/config/config.go`: Configuration loading and saving (`~/.config/frick/config.json`).
- `internal/gemini/client.go`: Connects to Google Gemini and Groq APIs. Includes custom JSON extraction parsing.
- `internal/shell/integration.go`: Generates Zsh/Bash integration scripts.
