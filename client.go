package main

import (
	"echo/utils"
	"fmt"
	"log"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

const HELP = `
Usage:
  echo [flags]

Flags:
  --mode string
        Mode of operation: send or receive (optional if using interactive mode)
  --local string  
        Local address to listen on (e.g. :9000) - for receive mode
  --remote string
        Remote peer address (e.g. 127.0.0.1:9001) - for send mode
  --file string
        File path to send (required in send mode)
  --help
        Show this help message and exit
  --bench
        Run benchmarking

Features:
  - QUIC protocol for reliable, fast transfer
  - Protobuf message serialization
  - MD5 checksum validation
  - Progress indicators
  - Automatic chunking (64KB chunks)

Interactive mode will start if no flags are provided.
`

const VERSION = 1

func main() {
	if err := mainEntry(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func mainEntry() error {
	cfg, err := utils.ParseFlags()
	if err != nil {
		return fmt.Errorf("flag parsing failed: %w", err)
	}

	if cfg.HelpMode {
		printHelpBox()
		return nil
	}

	if cfg.Mode == "" {
		handleSurveyMode(cfg)
	} else {
		if err := utils.ValidateFlags(cfg); err != nil {
			return fmt.Errorf("invalid input: %w", err)
		}
	}

	if err := RunPeer(cfg.LocalPort, cfg.RemoteAddr, cfg.FilePath, cfg.BenchMark, cfg.Mode); err != nil {
		return fmt.Errorf("run failed: %w", err)
	}

	return nil
}

func handleSurveyMode(cfg *utils.Config, opts ...survey.AskOpt) {
	var selectedMode string
	prompt := &survey.Select{
		Message: "Choose mode:",
		Options: []string{"Send a file", "Receive a file"},
	}
	survey.AskOne(prompt, &selectedMode, opts...)

	cfg.Mode = map[string]string{
		"Send a file":    "send",
		"Receive a file": "receive",
	}[selectedMode]

	blue := color.New(color.FgBlue).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Printf("\n%s Please choose your settings.\n", bold("CONFIGURATION"))

	if cfg.Mode == "receive" {
		survey.AskOne(&survey.Input{
			Message: fmt.Sprintf("%s Enter local address to listen on (e.g. :9000):", blue(">>")),
			Default: ":9000",
		}, &cfg.LocalPort, opts...)
	} else {
		survey.AskOne(&survey.Input{
			Message: fmt.Sprintf("%s Enter receiver's address (e.g. 127.0.0.1:9000):", blue(">>")),
		}, &cfg.RemoteAddr, opts...)

		survey.AskOne(&survey.Input{
			Message: fmt.Sprintf("%s Enter path to the file you want to send:", blue(">>")),
		}, &cfg.FilePath, opts...)
	}
}

func RunPeer(localAddr, remoteAddr, sendFile string, benchmark bool, mode string) error {
	if mode == "send" || sendFile != "" {
		return Send(sendFile, remoteAddr, benchmark)
	} else {
		return Receive(localAddr, benchmark)
	}
}

func printHelpBox() {
	boxColor := color.New(color.FgHiBlue, color.Bold)
	textColor := color.New(color.FgHiWhite)

	lines := strings.Split(HELP, "\n")
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	banner := color.New(color.FgGreen, color.Bold).Sprint(" Echo File Transfer (QUIC + Protobuf) ")

	boxColor.Println("╔" + strings.Repeat("═", maxWidth+2) + "╗")
	fmt.Printf("║%s%s║\n", banner, strings.Repeat(" ", maxWidth-len("Echo File Transfer (QUIC + Protobuf)")))
	boxColor.Println("╠" + strings.Repeat("═", maxWidth+2) + "╣")

	for _, line := range lines {
		textColor.Printf("║ %-*s ║\n", maxWidth, line)
	}

	boxColor.Println("╚" + strings.Repeat("═", maxWidth+2) + "╝")
}
