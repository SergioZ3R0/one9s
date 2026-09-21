package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/scabello/one9s/internal/client"
	"github.com/scabello/one9s/internal/config"
	"github.com/scabello/one9s/pkg/tui"
)

// Version is set at build time via -ldflags.
var Version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help", "help":
			printUsage()
			return
		case "-v", "--version", "version":
			fmt.Printf("one9s %s\n", Version)
			return
		case "vault":
			handleVault(os.Args[2:])
			return
		}
	}

	tui.SetVersion(Version)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "one9s: %v\n", err)
		os.Exit(1)
	}

	c := client.NewGOCA(cfg)

	p := tea.NewProgram(
		tui.NewRootModel(c),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "one9s: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`one9s - Terminal UI for OpenNebula cluster management

Usage:
  one9s                          Launch the TUI (default)
  one9s vault <command>          Manage encrypted credentials
  one9s -h, --help               Show this help
  one9s -v, --version            Show version

Vault Commands:
  one9s vault init               Create a new encrypted config file
  one9s vault encrypt            Encrypt an existing plain config file
  one9s vault decrypt            Decrypt and display the vault contents

Environment Variables:
  ONE_XMLRPC                     OpenNebula XML-RPC endpoint
  ONE_AUTH                       OpenNebula credentials (user:password)
  ONE_VAULT_PASS                 Vault password (skip prompt)

Config File:
  ~/.one9s/config                Plain text config (ONE_XMLRPC + ONE_AUTH)
  ~/.one9s/config.vault          Encrypted vault (AES-256-GCM)

Documentation:
  https://one9s.scszero.com/docs.html`)
}

func handleVault(args []string) {
	if len(args) == 0 {
		printVaultUsage()
		return
	}

	switch args[0] {
	case "init":
		vaultInit()
	case "encrypt":
		vaultEncrypt()
	case "decrypt":
		vaultDecrypt()
	default:
		printVaultUsage()
	}
}

func printVaultUsage() {
	fmt.Println("one9s vault - manage encrypted credentials")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  one9s vault init       Create a new encrypted config file")
	fmt.Println("  one9s vault encrypt    Encrypt an existing plain config file")
	fmt.Println("  one9s vault decrypt    Decrypt and display the vault contents")
}

func vaultInit() {
	vaultPath := config.VaultPath()
	if vaultPath == "" {
		fmt.Fprintln(os.Stderr, "error: cannot determine home directory")
		os.Exit(1)
	}

	// Check if vault already exists
	if _, err := os.Stat(vaultPath); err == nil {
		fmt.Fprintf(os.Stderr, "Vault already exists: %s\n", vaultPath)
		fmt.Fprintln(os.Stderr, "Delete it first or use 'one9s vault encrypt' to encrypt an existing config.")
		os.Exit(1)
	}

	fmt.Println("Create a new one9s vault configuration")
	fmt.Println()

	// Get vault password
	pass1 := readPassword("Vault password: ")
	pass2 := readPassword("Confirm password: ")
	if pass1 != pass2 {
		fmt.Fprintln(os.Stderr, "error: passwords do not match")
		os.Exit(1)
	}
	if pass1 == "" {
		fmt.Fprintln(os.Stderr, "error: password cannot be empty")
		os.Exit(1)
	}

	// Get credentials
	fmt.Print("OpenNebula user: ")
	user := readLine()
	pass := readPassword("OpenNebula password: ")

	fmt.Print("XML-RPC endpoint [http://localhost:2633/RPC2]: ")
	endpoint := readLine()
	if endpoint == "" {
		endpoint = "http://localhost:2633/RPC2"
	}

	if err := config.InitVault(vaultPath, pass1, user, pass, endpoint); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Vault created: %s\n", vaultPath)
	fmt.Println("Run 'one9s' to start (you will be prompted for the vault password).")
}

func vaultEncrypt() {
	vaultPath := config.VaultPath()
	configPath := strings.TrimSuffix(vaultPath, "config.vault") + "config"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Config file not found: %s\n", configPath)
		fmt.Fprintln(os.Stderr, "Create it first or use 'one9s vault init'.")
		os.Exit(1)
	}

	// Check if vault already exists
	if _, err := os.Stat(vaultPath); err == nil {
		fmt.Fprintf(os.Stderr, "Vault already exists: %s\n", vaultPath)
		fmt.Fprintln(os.Stderr, "Delete it first to re-encrypt.")
		os.Exit(1)
	}

	fmt.Print("Vault password: ")
	pass1 := readLine()
	fmt.Print("Confirm password: ")
	pass2 := readLine()
	if pass1 != pass2 {
		fmt.Fprintln(os.Stderr, "error: passwords do not match")
		os.Exit(1)
	}

	if err := config.EncryptConfig(vaultPath, configPath, pass1); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Encrypted: %s -> %s\n", configPath, vaultPath)
	fmt.Println("You can now delete the plain config file.")
}

func vaultDecrypt() {
	vaultPath := config.VaultPath()
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Vault not found: %s\n", vaultPath)
		os.Exit(1)
	}

	fmt.Println("Vault content:")
	pass := readPassword("Vault password: ")

	plain, err := config.DecryptVault(vaultPath, pass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(plain)
}

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func readPassword(prompt string) string {
	fmt.Fprint(os.Stderr, prompt)
	// On Linux, try to hide input with stty
	if _, err := exec.Command("stty", "-echo").StdinPipe(); err == nil {
		cmd := exec.Command("stty", "-echo")
		cmd.Stdin = os.Stdin
		_ = cmd.Run()
		defer func() {
			cmd = exec.Command("stty", "echo")
			cmd.Stdin = os.Stdin
			_ = cmd.Run()
		}()
	}
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
