package tui

import (
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	files     []table.Row
	directory string
	table     table.Model
	hexDeciph viewport.Model
	style     lipgloss.Style
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		hexData string
	)

	switch ms := msg.(type) {
	case tea.WindowSizeMsg:
		m.style.Width(ms.Width - 4)
		m.style.Height(ms.Height - 4)

		m.table.SetWidth(ms.Width - 4)
		m.table.SetHeight(ms.Height/2 - 4)
		m.hexDeciph.Width = ms.Width - 4
		m.hexDeciph.Height = ms.Height/2 - 4

	case tea.KeyMsg:
		switch ms.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			m.table.MoveUp(1)
		case "down":
			m.table.MoveDown(1)
		case "end":
			m.table.GotoBottom()
		case "home":
			m.table.GotoTop()
		case "enter":
			m.directory = filepath.Join(m.directory, m.table.SelectedRow()[1])

			stat, err := os.Stat(m.directory)
			if err != nil {
				hexData = err.Error()
				m.directory = filepath.Dir(m.directory)
				m.files = updateDirectory(m.directory)
				m.table.SetRows(m.files)
				m.hexDeciph.SetContent(hexData)
				m.table.SetCursor(0)
			}

			if stat.IsDir() {
				m.files = updateDirectory(m.directory)
				m.table.SetRows(m.files)
				m.table.SetCursor(0)
			} else {
				hexData, err = readSomeFileData(m.directory)
				if err != nil {
					hexData = err.Error()
				}
				m.directory = filepath.Dir(m.directory)
				m.hexDeciph.SetContent(hexData)
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	return m.style.Render(
		lipgloss.JoinVertical(
			lipgloss.Top,
			m.table.View(),
			m.hexDeciph.View(),
		),
	)
}

func updateDirectory(directory string) []table.Row {
	rows := []table.Row{}

	upperF, err := os.Stat("..")
	if err == nil {
		rows = append(rows, table.Row{
			upperF.Mode().String(),
			"..",
			upperF.ModTime().Local().Format("2006-01-02 15:04:05"),
		})
	}

	files, err := os.ReadDir(directory)
	if err != nil {
		return nil
	}

	for _, i := range files {
		// rows = append(rows, table.Row{
		// 	i.Type().String(),
		// 	i.Name(),
		// 	i.Type().Type().String(),
		// })
		stat, err := os.Stat(filepath.Join(directory, i.Name()))
		if err != nil {
			return nil
		}
		rows = append(rows, table.Row{
			stat.Mode().String(),
			i.Name(),
			stat.ModTime().Local().Format("2006-01-02 15:04:05"),
		})
	}

	return rows
}

func InitialMode() Model {
	var m = Model{}
	var err error
	if m.directory, err = os.Getwd(); err != nil {
		panic(err)
	}

	m.table = table.New(
		table.WithColumns(
			[]table.Column{
				{
					Title: "Permissions",
					Width: 30,
				},
				{
					Title: "File name",
					Width: 30,
				},
				{
					Title: "Last Modified",
					Width: 30,
				},
			},
		),
		table.WithFocused(true),
		table.WithRows(updateDirectory(m.directory)),
	)

	m.hexDeciph = viewport.New(10, 10)

	m.style = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Top).
		AlignVertical(lipgloss.Top).
		Border(lipgloss.DoubleBorder(), true)
	return m
}

func readSomeFileData(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return err.Error(), err
	}
	defer f.Close()

	var buffer [128]byte
	if _, err = io.ReadFull(f, buffer[:]); err != nil {
		return "", err
	}
	return string(buffer[:]), nil
}
