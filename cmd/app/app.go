package app

import (
	"buster_daemon/fmgo/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func Execute() {
	var appMod = tui.InitialMode()
	var p *tea.Program = tea.NewProgram(appMod)

	_, err := p.Run()
	if err != nil {
		panic(err)
	}
	print(appMod.View())
}
