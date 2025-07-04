package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Gateway interface {
	CreateSession() error
	AttachSession() error
	KillSession() error
	SplitWindowHorizontally(count int) error
	SetLayoutMainVertical() error
	WaitForPanesReady() error
	SelectPane(paneID string) error
	SetPaneInteractive(paneID string) error
	StartClaudeAgents(paneIDs []string) error
	FetchMainPaneID() (string, error)
	FetchNonMainPaneIDs() ([]string, error)
}

type gateway struct {
	commandFactory *commandFactory
}

func NewGateway(sessionName string) Gateway {
	return &gateway{
		commandFactory: NewCommandFactory(sessionName),
	}
}

func (g *gateway) CreateSession() error {
	if err := g.commandFactory.newSession().Run(); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}
	return nil
}

func (g *gateway) AttachSession() error {
	cmd := g.commandFactory.attachSession()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *gateway) KillSession() error {
	return g.commandFactory.killSession().Run()
}

func (g *gateway) SplitWindowHorizontally(count int) error {
	if count > 4 {
		return fmt.Errorf("split count %d exceeds maximum of 4", count)
	}

	for range count {
		cmd := g.commandFactory.splitWindowHorizontally()
		if err := runWithOutput(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (g *gateway) SetLayoutMainVertical() error {
	if err := runWithOutput(g.commandFactory.selectLayout("main-vertical")); err != nil {
		return err
	}

	mainPaneId, err := g.FetchMainPaneID()
	if err != nil {
		return fmt.Errorf("failed to find main pane: %w", err)
	}

	if err := runWithOutput(g.commandFactory.resizePaneX(mainPaneId, "25%")); err != nil {
		return err
	}

	nonMainPaneIds, err := g.FetchNonMainPaneIDs()
	if err != nil {
		return fmt.Errorf("failed to find non-main panes: %w", err)
	}

	for _, paneId := range nonMainPaneIds {
		if err := runWithOutput(g.commandFactory.resizePaneY(paneId, "25%")); err != nil {
			return err
		}
	}

	return nil
}

func (g *gateway) WaitForPanesReady() error {
	paneIds, err := g.listPanes()
	if err != nil {
		return fmt.Errorf("failed to list panes: %w", err)
	}

	for _, paneId := range paneIds {
		if err := g.waitForPaneReady(paneId); err != nil {
			return err
		}
	}

	return nil
}

func (g *gateway) SelectPane(paneID string) error {
	return runWithOutput(g.commandFactory.selectPane(paneID))
}

func (g *gateway) SetPaneInteractive(paneID string) error {
	interactiveScript := fmt.Sprintf(`{ while true; do echo -n "> "; read line; for pane in $(tmux list-panes -t %s -F "#{pane_id}:#{pane_left}" | grep -v ":0$" | cut -d: -f1); do tmux send-keys -t $pane "$line"; tmux send-keys -t $pane Enter; done; done; } 2>/dev/null`, g.commandFactory.sessionName)

	return runWithOutput(g.commandFactory.sendKeys(paneID, interactiveScript))
}

func (g *gateway) StartClaudeAgents(paneIDs []string) error {
	for _, paneId := range paneIDs {
		if err := runWithOutput(g.commandFactory.sendKeys(paneId, "claude")); err != nil {
			return err
		}
	}

	return nil
}

func (g *gateway) FetchMainPaneID() (string, error) {
	paneIds, err := g.listPanes()
	if err != nil {
		return "", fmt.Errorf("failed to list panes: %w", err)
	}

	if len(paneIds) == 0 {
		return "", fmt.Errorf("no panes found")
	}

	return paneIds[0], nil
}

func (g *gateway) FetchNonMainPaneIDs() ([]string, error) {
	mainPaneId, err := g.FetchMainPaneID()
	if err != nil {
		return nil, fmt.Errorf("failed to get main pane ID: %w", err)
	}

	allPaneIds, err := g.listPanes()
	if err != nil {
		return nil, fmt.Errorf("failed to list panes: %w", err)
	}

	var nonMainPaneIds []string
	for _, paneId := range allPaneIds {
		if paneId != mainPaneId {
			nonMainPaneIds = append(nonMainPaneIds, paneId)
		}
	}

	return nonMainPaneIds, nil
}

func (g *gateway) listPanes() ([]string, error) {
	output, err := g.commandFactory.listPanes("#{pane_id}").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("no panes found")
	}

	return lines, nil
}

func (g *gateway) waitForPaneReady(paneId string) error {
	for range 100 {
		time.Sleep(100 * time.Millisecond)

		output, err := g.commandFactory.capturePane(paneId).Output()
		if err != nil {
			continue
		}

		content := string(output)
		lines := strings.Split(content, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "$") || strings.Contains(line, "%") || strings.Contains(line, ">") || strings.Contains(line, "❯") || strings.Contains(line, "#") {
				if !strings.Contains(line, "echo") && len(line) > 0 {
					return nil
				}
			}
		}
	}

	return fmt.Errorf("pane %s not ready", paneId)
}

func runWithOutput(cmd *exec.Cmd) error {
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w, output: %s", err, string(output))
	}
	return nil
}
