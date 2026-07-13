package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/BriskAM/frick/internal/config"
	"github.com/BriskAM/frick/internal/gemini"
	"github.com/BriskAM/frick/internal/shell"
	"golang.org/x/term"
)

var drewLines int

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	switch subcommand {
	case "configure", "config":
		handleConfigure()
	case "init":
		handleInit()
	case "run":
		handleRun()
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: frick <subcommand> [options]")
	fmt.Println("Subcommands:")
	fmt.Println("  configure, config  Set/update your Gemini API Key")
	fmt.Println("  init               Output shell integration script")
	fmt.Println("  run                Suggest corrections for a failed command")
}

func handleConfigure() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your Gemini API Key (from Google AI Studio): ")
	key, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading key: %v\n", err)
		os.Exit(1)
	}
	key = strings.TrimSpace(key)
	if key == "" {
		fmt.Println("API Key cannot be empty.")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = &config.Config{}
	}
	cfg.ApiKey = key
	cfg.Model = config.DefaultModel

	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration saved successfully!")
}

func handleInit() {
	shell.PrintIntegrationScript()
}

func handleRun() {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	cmdArg := runCmd.String("cmd", "", "The failed command string")
	exitArg := runCmd.Int("exit", 0, "The exit code of the failed command")

	if len(os.Args) > 2 {
		_ = runCmd.Parse(os.Args[2:])
	}

	if *cmdArg == "" {
		fmt.Fprintln(os.Stderr, "Error: --cmd is required")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil || cfg.ApiKey == "" {
		fmt.Fprintln(os.Stderr, "\x1b[31;1mError: API key is not configured.\x1b[0m")
		fmt.Fprintln(os.Stderr, "Please run 'frick configure' first to set your API Key.")
		os.Exit(1)
	}

	// Double-Frick detection
	var lastSuggestion string
	var lastExit int

	lastSug, err := config.LoadLastSuggestion()
	if err == nil && lastSug != nil {
		// If the command we are currently trying to fix is equal to the last suggestion we ran,
		// and the current exit code is non-zero, then that suggested command failed.
		if lastSug.Corrected == *cmdArg && *exitArg != 0 {
			lastSuggestion = lastSug.Corrected
			lastExit = *exitArg
		}
	}

	// Fetch suggestions from Gemma
	suggestions, err := gemini.GetSuggestions(cfg.ApiKey, cfg.Model, cfg.ApiEndpoint, *cmdArg, *exitArg, lastSuggestion, lastExit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting suggestions: %v\n", err)
		os.Exit(1)
	}

	if len(suggestions) == 0 {
		os.Exit(0)
	}

	// Open /dev/tty for interactive input and output
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening /dev/tty: %v\n", err)
		os.Exit(1)
	}
	defer tty.Close()

	selectedIdx, err := runSelector(tty, suggestions)
	if err != nil {
		// Cancelled or interrupted
		os.Exit(0)
	}

	chosen := suggestions[selectedIdx]

	// Handle safety confirmation
	if chosen.SafetyLevel == "DANGER" {
		fmt.Fprintf(tty, "\n\x1b[31;1mDANGER:\x1b[0m This command is classified as dangerous.\n")
		fmt.Fprintf(tty, "Are you sure you want to run it? Type \x1b[31;1myes\x1b[0m (or \x1b[31;1my\x1b[0m) and press Enter: ")

		// Set tty back to normal input mode briefly to read string
		reader := bufio.NewReader(tty)
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(tty, "\nAborted.")
			os.Exit(0)
		}
		text = strings.TrimSpace(strings.ToLower(text))
		if text != "y" && text != "yes" {
			fmt.Fprintln(tty, "Aborted.")
			os.Exit(0)
		}
	}

	// Save this chosen suggestion for future double-frick checks
	_ = config.SaveLastSuggestion(&config.LastSuggestion{
		Command:   *cmdArg,
		Corrected: chosen.Command,
		ExitCode:  *exitArg,
	})

	// Print to stdout (not tty) so shell integration catches it
	fmt.Println(chosen.Command)
}

func runSelector(tty *os.File, suggestions []gemini.Suggestion) (int, error) {
	// Put tty in raw mode
	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return -1, err
	}
	defer term.Restore(int(tty.Fd()), oldState)

	// Hide cursor
	fmt.Fprint(tty, "\x1b[?25l")
	defer fmt.Fprint(tty, "\x1b[?25h")

	selected := 0
	numSuggestions := len(suggestions)

	drewLines = 0
	drawMenu(tty, suggestions, selected)

	buf := make([]byte, 3)
	for {
		n, err := tty.Read(buf)
		if err != nil {
			return -1, err
		}

		if n == 1 {
			b := buf[0]
			if b == '\r' || b == '\n' {
				// Enter pressed
				// Clear the menu before exiting
				clearMenu(tty)
				return selected, nil
			}
			if b == 3 { // Ctrl+C
				clearMenu(tty)
				return -1, fmt.Errorf("interrupted")
			}
			if b == 27 { // Esc
				clearMenu(tty)
				return -1, fmt.Errorf("cancelled")
			}
			if b == 'j' {
				selected = (selected + 1) % numSuggestions
				drawMenu(tty, suggestions, selected)
			}
			if b == 'k' {
				selected = (selected - 1 + numSuggestions) % numSuggestions
				drawMenu(tty, suggestions, selected)
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == '[' {
			// Escape sequence for arrow key
			if buf[2] == 'A' { // Up
				selected = (selected - 1 + numSuggestions) % numSuggestions
				drawMenu(tty, suggestions, selected)
			} else if buf[2] == 'B' { // Down
				selected = (selected + 1) % numSuggestions
				drawMenu(tty, suggestions, selected)
			}
		}
	}
}

func drawMenu(tty *os.File, suggestions []gemini.Suggestion, selected int) {
	if drewLines > 0 {
		for i := 0; i < drewLines; i++ {
			fmt.Fprint(tty, "\x1b[F\x1b[K")
		}
	}

	fmt.Fprintln(tty, "\x1b[35m?\x1b[0m Select correction (Enter to run, Ctrl+C to cancel):")
	drewLines = 1

	for i, s := range suggestions {
		prefix := "  "
		if i == selected {
			prefix = "\x1b[36m> \x1b[0m"
		}

		var cmdStr string
		switch s.SafetyLevel {
		case "SAFE":
			if i == selected {
				cmdStr = fmt.Sprintf("\x1b[1;32m%s\x1b[0m", s.Command)
			} else {
				cmdStr = fmt.Sprintf("\x1b[32m%s\x1b[0m", s.Command)
			}
		case "WARNING":
			if i == selected {
				cmdStr = fmt.Sprintf("\x1b[1;33m%s\x1b[0m", s.Command)
			} else {
				cmdStr = fmt.Sprintf("\x1b[33m%s\x1b[0m", s.Command)
			}
		case "DANGER":
			if i == selected {
				cmdStr = fmt.Sprintf("\x1b[1;31m%s\x1b[0m", s.Command)
			} else {
				cmdStr = fmt.Sprintf("\x1b[31m%s\x1b[0m", s.Command)
			}
		default:
			cmdStr = s.Command
		}

		fmt.Fprintf(tty, "%s%s\n", prefix, cmdStr)
		drewLines++
	}
}

func clearMenu(tty *os.File) {
	if drewLines > 0 {
		for i := 0; i < drewLines; i++ {
			fmt.Fprint(tty, "\x1b[F\x1b[K")
		}
		drewLines = 0
	}
}
