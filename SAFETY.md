# Safety and Security Report - frick

This document provides a thorough analysis of the security architecture, threat model, and mitigation controls implemented in the `frick` command-line utility. 

Because `frick` interacts directly with your shell command execution pipeline, maintaining strict security boundaries is a primary goal of the project.

---

## 1. Security Architecture and Execution Model

Unlike traditional shell automation utilities that may run corrected commands in background sub-processes, `frick` separates **suggestion generation** from **execution**.

```
+--------------------------------------------------------+
|                      Shell Session                     |
|  1. Typoed command fails                              |
|  2. User invokes `frick` function                     |
+---------------------------+----------------------------+
                            |
                            | (Launches Binary)
                            v
+--------------------------------------------------------+
|                     frick Binary                       |
|  3. Reads local config (0600 permissions)              |
|  4. Calls remote API (HTTPS) to get corrections        |
|  5. Displays colored interactive selection menu       |
|  6. Performs local Safety Override keyword scan        |
|  7. (If DANGER) Forces raw TTY y/yes confirmation      |
|  8. Outputs selected command to STDOUT and exits       |
+---------------------------+----------------------------+
                            |
                            | (Returns selection)
                            v
+--------------------------------------------------------+
|                      Shell Session                     |
|  9. Active shell evaluates STDOUT via `eval`           |
+--------------------------------------------------------+
```

This model ensures that the binary itself has **zero execution privilege**. It cannot spawn shells or run system processes directly. All executions happen transparently in the user's active shell session under their full visibility.

---

## 2. Threat Modeling & Vulnerability Analysis

### Threat A: Arbitrary Command Execution via Prompt Injection
*   **Description**: An attacker manipulates local files, directory names, or environment variables to cause the remote AI model (Gemini or Qwen) to return a malicious command (e.g. `rm -rf /` or unauthorized network curls).
*   **Severity**: Critical
*   **Implemented Mitigations**:
    1.  **No Automatic Execution**: Suggestions are only executed after active user confirmation (pressing Enter).
    2.  **Safety Categorization**: Commands are audited for safety levels:
        - `SAFE`: Non-modifying commands (e.g., `git status`, `ls`).
        - `WARNING`: Modifying commands (e.g., `npm install`, `git commit`).
        - `DANGER`: Highly destructive commands (e.g., `rm`, `sudo`, `dd`).
    3.  **Explicit Danger Confirms**: Destructive commands bypass simple keypresses and force the user to type `yes` or `y` in a raw TTY prompt.
    4.  **Local Safety Overrides**: Users can configure the `safety_overrides` map in `~/.config/frick/config.json` to locally force specific keywords (such as `production`, `deploy`, or internal command prefixes) to be treated as `DANGER`, bypassing the AI model's estimation entirely.

### Threat B: Local Credential Extraction (API Keys)
*   **Description**: A malicious local user or a restricted process on the same machine reads the Gemini or Groq API keys from the configuration directory.
*   **Severity**: High
*   **Implemented Mitigations**:
    - **Owner-Only File Permissions**: The `SaveConfig` function writes the configuration to `~/.config/frick/config.json` using Unix file mode `0600` (Read/Write owner only). This blocks other system users from reading the configuration file.

### Threat C: Terminal State Hijacking (Raw Mode Lock)
*   **Description**: The program enters raw terminal mode to read keyboard input (arrow keys) but crashes or gets interrupted, leaving the user's terminal session unusable.
*   **Severity**: Medium
*   **Implemented Mitigations**:
    - **Deferred Recovery**: Raw mode configuration immediately registers a deferred `Restore` hook:
      ```go
      oldState, err := term.MakeRaw(int(tty.Fd()))
      if err == nil {
          defer term.Restore(int(tty.Fd()), oldState)
      }
      ```
      This guarantees that even in the event of an unexpected application panic, the terminal state is restored to normal mode.

### Threat D: Memory Exploits and Buffer Overflows
*   **Description**: Malformed command outputs or malicious terminal input strings trigger buffer overflows, allowing arbitrary code execution within the utility memory space.
*   **Severity**: Low
*   **Implemented Mitigations**:
    - **Memory-Safe Runtime**: Written entirely in Go, which implements automated garbage collection, strict type checks, boundary-checks on slices/arrays, and prohibits raw pointer arithmetic by default.

---

## 3. Best Practices for Users
- **Do not run as root/sudo**: Avoid running shell sessions or setting up shell aliases as `root`. Run `frick` in user space to maintain standard security separation.
- **Maintain your Safety Overrides**: Add any sensitive command keywords (like internal databases, deployment scripts, or remote server connections) to your local `safety_overrides` mapping.

---

## 4. Reporting Vulnerabilities
If you discover a security vulnerability in `frick`, please do not open a public issue. Instead, report it privately by contacting the project maintainers directly.
