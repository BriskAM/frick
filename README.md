# frick

`frick` is a lightning-fast command-line utility inspired by `thefuck`. Written in Go, it runs with sub-millisecond start times, leverages Gemini API / Groq API models to fix mistyped terminal commands, and integrates with Zsh/Bash shells.

## Features

- **Fast**: Written in Go with minimal overhead. Startup time is under 5ms.
- **AI-Powered**: Uses Gemini/Groq models to analyze typos, directory context, and active environment setups.
- **Safety Evaluation Layer**: Color-coded warnings for suggested commands:
  - **Safe** (e.g. `git status`): Standard informational commands. Press **Enter** to run instantly.
  - **Warning** (e.g. `npm install`): Modifying commands. Press **Enter** to run instantly.
  - **Danger** (e.g. `rm -rf`): Highly destructive operations. Requires typing **`y`** or **`yes`** then **Enter** to execute.
- **Directory & Env Context**: Automatically injects directory listings and virtualenv states so that the model suggests the right filenames/module paths.
- **Double-Frick / Self-Correction**: If a correction fails and you run `frick` again immediately, it detects the failure and suggests a different approach.
- **Clean History**: Replaces the failed command and the `frick` call in your active shell history with the corrected command.

---

## Installation

### 1. Fast Installation (macOS & Linux)
Install instantly via curl:
```bash
curl -fsSL https://raw.githubusercontent.com/BriskAM/frick/main/install.sh | bash
```

### 2. Manual Build from Source
If you prefer to build manually:
```bash
git clone https://github.com/BriskAM/frick.git
cd frick
go build -o frick cmd/frick/main.go
mv frick /usr/local/bin/
```

### 3. Configure API Key
Configure your Gemini or Groq API key:
```bash
frick configure
```

### 4. Custom API Endpoint (Optional)
If you want to use a proxy, local gateway (like Ollama), or enterprise custom endpoint, edit `~/.config/frick/config.json` and add `api_endpoint`:
```json
{
  "api_key": "YOUR_API_KEY",
  "model": "gemini-3.1-flash-lite",
  "api_endpoint": "https://custom-gateway.local"
}
```

### 5. Setup Shell Integration
Add the following line to your `~/.zshrc` (for Zsh) or `~/.bashrc` (for Bash):
```bash
eval "$(frick init)"
```
Then restart your shell or run `source ~/.zshrc` / `source ~/.bashrc`.

---

## Usage

Type any wrong command, and then simply type `frick`:

```bash
$ gti status
zsh: command not found: gti

$ frick
? Select correction (Enter to run, Ctrl+C to cancel):
> git status [SAFE]
  git status -s [SAFE]
  git branch [SAFE]
```

Use the **Up/Down arrow keys** to cycle through suggestions, and press **Enter** to run the selected command.

---

## License
MIT License. See [LICENSE](LICENSE) for details.
