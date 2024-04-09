package mountviews

import (
	"buster_daemon/fmgo/internal/mounts"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MountModel struct {
	mounts []table.Row
	table  table.Model
	style  lipgloss.Style
}

func (m MountModel) Init() tea.Cmd {
	return nil
}

func (m MountModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch ms := msg.(type) {
	case tea.WindowSizeMsg:
		m.style.Height(ms.Height)
		m.style.Width(ms.Width)

	case tea.KeyMsg:
		switch ms.String() {
		case tea.KeyCtrlC.String(),
			"q", tea.KeyEsc.String():
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MountModel) View() string {
	return m.style.Render(
		m.table.View(),
	)
}

func New() (MountModel, error) {
	var m MountModel
	m.style = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		AlignHorizontal(lipgloss.Center).
		AlignVertical(lipgloss.Left)

	m.table = table.New(
		table.WithColumns([]table.Column{
			{
				Title: "Mount point",
				Width: lipgloss.Style.GetMaxWidth(m.style),
			},
		}),
	)

	mountList, err := mounts.GetMounts()
	if err != nil {
		return MountModel{}, err
	}

	for _, l := range mountList {
		m.mounts = append(
			m.mounts,
			table.Row{
				l,
			})
	}
	m.table.SetRows(m.mounts)

	return m, nil
}
