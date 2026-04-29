package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

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

	// Terminal handling
	initialState, _ := term.GetState(int(os.Stdin.Fd()))
	handleExit := func() {
		term.Restore(int(os.Stdin.Fd()), initialState)
		exec.Command("stty", "sane").Run()
	}

	options := append(getStyleOptions(),
		prompt.OptionPrefix("cerebro> "),
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

	select {
	case <-operator.Wait():
		log.Println("connection closed.")
	case <-promptExitCh:
		log.Println("session ended.")
	}

	handleExit()
	log.Println("exiting...")
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

		prompt.OptionShowCompletionAtStart(),
	}
}
