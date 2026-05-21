package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/c-bata/go-prompt"
	"golang.org/x/term"
)

func main() {
	const defaultWSURL = "wss://nitronode-sandbox.yellow.org/v1/ws"

	log.SetFlags(0)
	log.SetPrefix("cerebro: ")
	log.SetOutput(os.Stderr)

	// Resolve config directory: explicit env > legacy "clearnode-cli" > "cerebro"
	configDir := os.Getenv("NITRONODE_CLI_CONFIG_DIR")
	if configDir == "" {
		userConfDir, err := os.UserConfigDir()
		if err != nil {
			log.Fatalf("failed to get user config directory: %v", err)
		}

		legacyDir := filepath.Join(userConfDir, "clearnode-cli")
		newDir := filepath.Join(userConfDir, "cerebro")

		if info, err := os.Stat(legacyDir); err == nil && info.IsDir() {
			configDir = legacyDir
			fmt.Printf("WARNING: Using legacy config directory %s\n", legacyDir)
			fmt.Printf("         Please rename it to %s\n", newDir)
		} else {
			configDir = newDir
		}
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Fatalf("failed to create config directory: %v", err)
	}

	// Initialize storage
	storagePath := filepath.Join(configDir, "config.db")
	store, err := NewStorage(storagePath)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	// Determine WebSocket URL: CLI arg > stored > default
	var wsURL string
	if len(os.Args) >= 2 {
		wsURL = os.Args[1]
	} else if stored, err := store.GetWSURL(); err == nil {
		wsURL = stored
	} else {
		wsURL = defaultWSURL
	}

	// Create operator
	operator, err := NewOperator(wsURL, configDir, store)
	if err != nil {
		log.Fatalf("failed to create operator: %v", err)
	}

	fmt.Println("Cerebro - Nitrolite SDK Development Tool")
	fmt.Printf("Connected to: %s\n", wsURL)
	fmt.Printf("Config directory: %s\n", configDir)
	fmt.Println("\nType 'help' for available commands or 'exit' to quit")

	// Terminal handling. go-prompt switches the tty into raw mode and emits
	// bracketed-paste / alternate-screen / mouse-tracking escapes. Restoring
	// the termios is not enough — the terminal also needs the matching
	// disable sequences or the next program in the same shell tab inherits
	// broken paste behaviour and an unresponsive Ctrl-C.
	initialState, _ := term.GetState(int(os.Stdin.Fd()))
	handleExit := func() {
		term.Restore(int(os.Stdin.Fd()), initialState)
		exec.Command("stty", "sane").Run()
		// Disable bracketed paste, show cursor, leave alt screen, mouse off,
		// then wipe any leftover completion-menu / ghost-prompt rows that
		// go-prompt leaves on the screen during its teardown.
		fmt.Fprint(os.Stdout, "\x1b[?2004l\x1b[?25h\x1b[?1049l\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l")
		// go-prompt leaves on the screen: the redrawn ghost prompt line
		// (rendered between Executor return and ExitChecker firing) plus up
		// to 6 rows reserved for the completion menu (default maxSuggestion
		// = 6 in go-prompt v0.2.6). Cursor is positioned somewhere below
		// them. Step cursor up enough rows to land at or above the ghost
		// prompt, then erase from there down. Over-erasing into the user's
		// "cerebro> exit" line is acceptable — the farewell printed below
		// preserves the meaning.
		fmt.Fprint(os.Stdout, "\x1b[7A\r\x1b[0J")
	}

	// Catch SIGINT / SIGTERM / SIGHUP so abnormal termination still restores
	// the terminal. The in-prompt Ctrl-C keybind below handles the normal
	// "exit by Ctrl-C" path; this handler covers shell-level kills.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigCh
		handleExit()
		os.Exit(130)
	}()

	options := append(getStyleOptions(),
		prompt.OptionPrefix("cerebro> "),
		// Make go-prompt exit its Run() loop when the user types "exit" so
		// its own tty teardown runs before main prints anything below the
		// prompt. Without this, exit racing with the prompt redraw leaves
		// completion suggestions splattered over the final output.
		prompt.OptionSetExitCheckerOnInput(func(in string, breakline bool) bool {
			return breakline && strings.TrimSpace(in) == "exit"
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(_ *prompt.Buffer) {
				log.Println("exiting Cerebro")
				handleExit()
				os.Exit(0)
			},
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlD,
			Fn:  func(_ *prompt.Buffer) {},
		}),
	)

	p := prompt.New(
		operator.Execute,
		operator.Complete,
		options...,
	)

	promptExitCh := make(chan struct{})
	go func() {
		p.Run()
		close(promptExitCh)
	}()

	var farewell string
	select {
	case <-operator.Wait():
		farewell = "connection closed."
	case <-promptExitCh:
		farewell = "session ended."
	}

	// Restore the terminal first so the farewell prints below any leftover
	// prompt artefacts in normal mode, not interleaved with go-prompt's
	// final redraw.
	handleExit()
	log.Println(farewell)
}

func getStyleOptions() []prompt.Option {
	return []prompt.Option{
		prompt.OptionTitle("Cerebro"),
		prompt.OptionPrefixTextColor(prompt.Green),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),

		prompt.OptionSuggestionTextColor(prompt.White),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),

		prompt.OptionDescriptionTextColor(prompt.Black),
		prompt.OptionDescriptionBGColor(prompt.Cyan),

		prompt.OptionSelectedSuggestionTextColor(prompt.Black),
		prompt.OptionSelectedSuggestionBGColor(prompt.Green),

		prompt.OptionSelectedDescriptionTextColor(prompt.White),
		prompt.OptionSelectedDescriptionBGColor(prompt.DarkBlue),
	}
}
