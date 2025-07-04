package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sh0623k/claude-mux/pkg/gateways/tmux"
)

const sessionName = "manager-session"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	tmuxGateway := tmux.NewGateway(sessionName)

	if err := tmuxGateway.CreateSession(); err != nil {
		return err
	}
	defer func() {
		tmuxGateway.KillSession()
	}()

	if err := tmuxGateway.SplitWindowHorizontally(4); err != nil {
		return fmt.Errorf("failed to split window: %w", err)
	}

	if err := tmuxGateway.SetLayoutMainVertical(); err != nil {
		return fmt.Errorf("failed to set layout: %w", err)
	}

	if err := tmuxGateway.WaitForPanesReady(); err != nil {
		return fmt.Errorf("failed to wait for pane ready: %w", err)
	}

	mainPaneID, err := tmuxGateway.FetchMainPaneID()
	if err != nil {
		return fmt.Errorf("failed to fetch main pane ID: %w", err)
	}

	if err := tmuxGateway.SelectPane(mainPaneID); err != nil {
		return fmt.Errorf("failed to select main pane: %w", err)
	}

	if err := tmuxGateway.SetPaneInteractive(mainPaneID); err != nil {
		return fmt.Errorf("failed to set main pane interactive: %w", err)
	}

	nonMainPaneIDs, err := tmuxGateway.FetchNonMainPaneIDs()
	if err != nil {
		return fmt.Errorf("failed to fetch non-main pane IDs: %w", err)
	}

	if err := tmuxGateway.StartClaudeAgents(nonMainPaneIDs); err != nil {
		return fmt.Errorf("failed to start claude agents: %w", err)
	}

	setupSignalHandler(tmuxGateway)

	return tmuxGateway.AttachSession()
}

func setupSignalHandler(tmuxGateway tmux.Gateway) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		tmuxGateway.KillSession()
		os.Exit(0)
	}()
}
