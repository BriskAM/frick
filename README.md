# frick 🦺

`frick` is a lightning-fast command-line utility inspired by `thefuck`. Written in Go, it runs with sub-millisecond start times, leverages **Gemma 4** (`gemma-4-31b-it`) via the Gemini API to fix mistyped terminal commands, and integrates with Zsh/Bash shells.

## Features

- 🏎️ **Fast**: Written in Go with minimal overhead. Startup time is under 5ms.
- 🧠 **AI-Powered**: Uses Gemma 4 to analyze typos, directory context, and active environment setups.
- 🛡️ **Safety Evaluation Layer**: Color-coded warnings for suggested commands:
  - 🟢 **Safe** (e.g. `git status`): Standard informational commands. Press **Enter** to run instantly.
  - 🟡 **Warning** (e.g. `npm install`): Modifying commands. Press **Enter** to run instantly.
  - 🔴 **Danger** (e.g. `rm -rf`): Highly destructive operations. Requires typing **`y`** or **`yes`** then **Enter** to execute.
- 📂 **Directory & Env Context**: Automatically injects directory listings and virtualenv states so that Gemma suggests the right filenames/module paths.
- 🔄 **Double-Frick / Self-Correction**: If a correction fails and you run `frick` again immediately, it detects the failure and suggests a different approach.
- 🧹 **Clean History**: Replaces the failed command and the `frick` call in your active shell history with the corrected command.

---

## Installation

### 1. Build from Source
Ensure you have Go 1.21+ installed:
```bash
git clone https://github.com/BriskAM/frick.git
cd frick
go build -o frick cmd/frick/main.go
# Move it to your PATH
mv frick /usr/local/bin/
```

### 2. Configure API Key
Get a free Gemini API key from [Google AI Studio](https://aistudio.google.com/) and configure it:
```bash
frick configure
```

### 3. Setup Shell Integration
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
