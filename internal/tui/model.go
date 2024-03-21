package tui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	DELETE_STATUS_OFF = iota
	DELETE_STATUS_ON
)

const (
	CURRENT_DIRECTORY = "Current directory: %s"
	CONFIRM_DELETE    = "Are you sure want to delete the %s?"
)

type Model struct {
	files           []table.Row
	directory       string
	table           table.Model
	hexDeciph       viewport.Model
	deleteFDirState int8
	style           lipgloss.Style
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
		m.table.UpdateViewport()

		m.hexDeciph.Width = ms.Width - 4
		m.hexDeciph.Height = ms.Height/2 - 4

	case tea.KeyMsg:
		switch ms.String() {
		case "c":
			m.hexDeciph.SetContent("")
		case "delete":
			if m.deleteFDirState == DELETE_STATUS_ON {
				err := os.Remove(
					filepath.Join(
						m.directory,
						m.table.SelectedRow()[1],
					))

				if err != nil {
					m.hexDeciph.SetContent(err.Error())
				}
				m.table.SetRows(updateDirectory(m.directory))
				m.table.SetCursor(0)
				m.deleteFDirState = DELETE_STATUS_OFF
				return m, nil
			}

			if m.deleteFDirState == DELETE_STATUS_OFF {
				m.deleteFDirState = DELETE_STATUS_ON
				m.directory = filepath.Join(m.directory, m.table.SelectedRow()[1])
				return m, nil
			}
		case tea.KeyEsc.String(), tea.KeyEscape.String():
			if m.deleteFDirState == DELETE_STATUS_ON {
				m.deleteFDirState = DELETE_STATUS_OFF
				m.directory = filepath.Dir(m.directory)
			}
			return m, nil
		case tea.KeyCtrlC.String(), "q":
			return m, tea.Quit
		case tea.KeyUp.String():
			m.table.MoveUp(1)
		case tea.KeyDown.String():
			m.table.MoveDown(1)
		case tea.KeyEnd.String():
			m.table.GotoBottom()
		case tea.KeyHome.String():
			m.table.GotoTop()
		case tea.KeyPgDown.String():
			m.table.MoveDown(lipgloss.Height(m.table.View()))
		case tea.KeyPgUp.String():
			m.table.MoveUp(lipgloss.Height(m.table.View()))
		case tea.KeyBackspace.String():
			m.directory = filepath.Dir(m.directory)
			m.files = updateDirectory(m.directory)
			m.table.SetRows(updateDirectory(m.directory))
			m.table.SetCursor(0)
		case tea.KeyEnter.String(), "return":
			m.hexDeciph.SetContent("")
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
				hexData, err = readSomeFileData(m.directory, 1)
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

			fmt.Sprintf(
				func() string {
					if m.deleteFDirState == DELETE_STATUS_ON {
						return CONFIRM_DELETE
					}
					return CURRENT_DIRECTORY
				}(),
				m.directory),

			lipgloss.PlaceHorizontal(
				lipgloss.Width(
					m.table.View(),
				),
				lipgloss.Center,
				"Decoded View:",
				lipgloss.WithWhitespaceChars(lipgloss.BlockBorder().Bottom),
			),
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
			"",
			upperF.ModTime().Local().Format("2006-01-02 15:04:05"),
		})
	}

	files, err := os.ReadDir(directory)
	if err != nil {
		return nil
	}

	for _, i := range files {
		stat, err := os.Stat(filepath.Join(directory, i.Name()))
		if err != nil {
			continue
		}
		rows = append(rows, table.Row{
			stat.Mode().String(),
			i.Name(),
			fmt.Sprintf("%d", stat.Size()),
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
					Width: 12,
				},
				{
					Title: "File name",
					Width: 30,
				},
				{
					Title: "File size",
					Width: 25,
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

	m.hexDeciph = viewport.New(5, 5)
	m.deleteFDirState = 0

	m.style = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Top).
		AlignVertical(lipgloss.Top).
		Border(lipgloss.DoubleBorder(), true)

	return m
}

func readSomeFileData(path string, mode int8) (string, error) {
	var resString string
	f, err := os.Open(path)
	if err != nil {
		return err.Error(), err
	}
	defer f.Close()

	var buffer [128]byte
	if _, err = io.ReadFull(f, buffer[:]); err != nil {
		return "", err
	}
	switch mode {
	case 1:
		for _, bf := range buffer {
			resString += fmt.Sprintf("%x", bf)
			resString += " "

		}
	case 2:
		resString = fmt.Sprintf("%s", buffer)
	default:
		resString = fmt.Sprintf("%s", buffer)
	}

	return resString, nil
}
