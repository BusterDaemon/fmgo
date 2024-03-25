package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//go:noinline
func getMBSize(x float32) float32

const (
	NORMAL_STATUS = iota
	DELETE_STATUS
	READ_FILE
	RENAME_FILE
	VIEW_MOUNTPOINTS
)

const (
	CURRENT_DIRECTORY = "Current directory: %s"
	CONFIRM_DELETE    = "Are you sure want to delete the %s?"
)

type Model struct {
	files      []table.Row
	directory  string
	textBar    string
	table      table.Model
	fNameInput textinput.Model
	uiStatus   uint8
	fileData   viewport.Model
	style      lipgloss.Style
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

		switch m.uiStatus {
		case NORMAL_STATUS:
			m.table.SetWidth(ms.Width - 4)
			m.table.SetHeight(ms.Height - 4)
			m.table.UpdateViewport()
		case READ_FILE:
			m.fileData.Width = ms.Width - 4
			m.fileData.Height = ms.Height - 4
		}
	case tea.KeyMsg:
		switch ms.String() {
		case "delete":
			if m.uiStatus == DELETE_STATUS {
				err := os.Remove(
					m.directory,
				)
				if err != nil {
					m.textBar = err.Error()
				}

				m.textBar = CURRENT_DIRECTORY
				m.directory = filepath.Dir(
					m.directory,
				)
				m.table.SetRows(updateDirectory(m.directory))
				m.table.SetCursor(0)
				m.uiStatus = NORMAL_STATUS
				return m, nil
			}

			if m.uiStatus == NORMAL_STATUS {
				m.uiStatus = DELETE_STATUS
				m.directory = filepath.Join(m.directory, m.table.SelectedRow()[1])
				m.textBar = CONFIRM_DELETE
				return m, nil
			}
		case tea.KeyEsc.String(), tea.KeyEscape.String():
			if m.uiStatus == DELETE_STATUS {
				m.uiStatus = NORMAL_STATUS
				m.directory = filepath.Dir(m.directory)
				m.textBar = CURRENT_DIRECTORY
			}

			if m.uiStatus == READ_FILE {
				m.table.Focus()
				m.uiStatus = NORMAL_STATUS
			}

			return m, nil
		case tea.KeyCtrlC.String(), "q":
			return m, tea.Quit
		case tea.KeyUp.String():
			if m.table.Focused() {
				m.table.MoveUp(1)
			} else {
				m.fileData.LineUp(1)
			}
		case tea.KeyDown.String():
			if m.table.Focused() {
				m.table.MoveDown(1)
			} else {
				m.fileData.LineDown(1)
			}
		case tea.KeyEnd.String():
			if m.table.Focused() {
				m.table.GotoBottom()
			} else {
				m.fileData.GotoBottom()
			}
		case tea.KeyHome.String():
			if m.table.Focused() {
				m.table.GotoTop()
			} else {
				m.fileData.GotoTop()
			}
		case tea.KeyPgDown.String():
			if m.table.Focused() {
				m.table.MoveDown(lipgloss.Height(m.table.View()))
			}
		case tea.KeyPgUp.String():
			if m.table.Focused() {
				m.table.MoveUp(lipgloss.Height(m.table.View()))
			}
		case tea.KeyBackspace.String():
			if m.table.Focused() {
				m.directory = filepath.Dir(m.directory)
				m.files = updateDirectory(m.directory)
				m.table.SetRows(updateDirectory(m.directory))
				m.table.SetCursor(0)
			}
		case tea.KeyEnter.String(), "return":
			m.directory = filepath.Join(m.directory, m.table.SelectedRow()[1])

			stat, err := os.Stat(m.directory)
			if err != nil {
				hexData = err.Error()
				m.directory = filepath.Dir(m.directory)
				m.files = updateDirectory(m.directory)
				m.table.SetRows(m.files)
				m.fileData.SetContent(hexData)
				m.table.SetCursor(0)
			}

			if stat.IsDir() {
				m.files = updateDirectory(m.directory)
				m.table.SetRows(m.files)
				m.table.SetCursor(0)
			} else {
				m.uiStatus = READ_FILE
				m.table.Blur()
				hexData, err = readSomeFileData(m.directory, 2)
				if err != nil {
					hexData = err.Error()
				}
				m.directory = filepath.Dir(m.directory)
				m.fileData.SetContent(hexData)
				m.fileData.Width = lipgloss.Width(m.table.View())
				m.fileData.Height = lipgloss.Height(m.table.View())
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	return m.style.Render(
		func() string {
			if m.uiStatus == NORMAL_STATUS ||
				m.uiStatus == DELETE_STATUS {
				return lipgloss.JoinVertical(
					lipgloss.Top,
					m.table.View(),
					fmt.Sprintf(
						m.textBar,
						m.directory),
				)
			} else {
				return lipgloss.JoinHorizontal(
					lipgloss.Center,
					m.fileData.View(),
				)
			}
		}(),
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
		stat, err := os.Stat(
			filepath.Join(directory, i.Name()),
		)
		if err != nil {
			continue
		}
		rows = append(
			rows, table.Row{
				stat.Mode().String(),
				i.Name(),
				func() string {
					if !stat.IsDir() {
						return fmt.Sprintf("%.2f MB",
							getMBSize(
								float32(stat.Size()),
							),
						)
					}
					return ""
				}(),
				stat.ModTime().Local().Format("2006-01-02 15:04:05"),
			},
		)
	}

	return rows
}

func InitialMode() Model {
	var m = Model{}
	var err error
	m.uiStatus = NORMAL_STATUS
	m.textBar = CURRENT_DIRECTORY
	m.fileData = viewport.New(10, 10)
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

	m.fNameInput = textinput.New()

	m.style = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Top).
		AlignVertical(lipgloss.Top).
		Border(lipgloss.DoubleBorder(), true)

	return m
}

func readSomeFileData(path string, mode int8) (string, error) {
	var resString string
	f, err := os.ReadFile(path)
	if err != nil {
		return err.Error(), err
	}

	switch mode {
	case 1:
		for _, bf := range f {
			resString += fmt.Sprintf("%x", bf)
			resString += " "

		}
	case 2:
		resString = string(f)
	default:
		resString = string(f)
	}

	return resString, nil
}
