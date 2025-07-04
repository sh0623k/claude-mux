package tmux

import (
	"os/exec"
)

type commandFactory struct {
	sessionName string
}

func NewCommandFactory(sessionName string) *commandFactory {
	return &commandFactory{
		sessionName: sessionName,
	}
}

func (f *commandFactory) create(args ...string) *exec.Cmd {
	return exec.Command("tmux", args...)
}

func (f *commandFactory) newSession() *exec.Cmd {
	return f.create("new-session", "-d", "-s", f.sessionName)
}

func (f *commandFactory) attachSession() *exec.Cmd {
	return f.create("attach-session", "-t", f.sessionName)
}

func (f *commandFactory) killSession() *exec.Cmd {
	return f.create("kill-session", "-t", f.sessionName)
}

func (f *commandFactory) splitWindowHorizontally() *exec.Cmd {
	return f.create("split-window", "-t", f.sessionName, "-h")
}

func (f *commandFactory) selectLayout(layout string) *exec.Cmd {
	return f.create("select-layout", "-t", f.sessionName, layout)
}

func (f *commandFactory) resizePaneX(paneId string, size string) *exec.Cmd {
	return f.create("resize-pane", "-t", paneId, "-x", size)
}

func (f *commandFactory) resizePaneY(paneId string, size string) *exec.Cmd {
	return f.create("resize-pane", "-t", paneId, "-y", size)
}

func (f *commandFactory) selectPane(paneId string) *exec.Cmd {
	return f.create("select-pane", "-t", paneId)
}

func (f *commandFactory) listPanes(format string) *exec.Cmd {
	return f.create("list-panes", "-t", f.sessionName, "-F", format)
}

func (f *commandFactory) capturePane(paneId string) *exec.Cmd {
	return f.create("capture-pane", "-t", paneId, "-p")
}

func (f *commandFactory) sendKeys(paneId string, keys string) *exec.Cmd {
	return f.create("send-keys", "-t", paneId, keys, "Enter")
}
