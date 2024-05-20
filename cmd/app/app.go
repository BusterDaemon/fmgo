package app

import (
	"buster_daemon/fmgo/internal/splash"
	"buster_daemon/fmgo/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func Execute() {
	var splash = splash.New()
	var s *tea.Program = tea.NewProgram(splash)
	s.Run()

	var appMod = tui.InitialMode()
	var p *tea.Program = tea.NewProgram(appMod)

	_, err := p.Run()
	if err != nil {
		panic(err)
	}
}
