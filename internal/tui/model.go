package tui

import (
	"buster_daemon/fmgo/internal/mounts"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

//go:noinline
func getMBSize(x float32) (float32, int)

const (
	NORMAL_STATUS = iota
	DELETE_STATUS
	READ_FILE
	RENAME_FILE
	CREATE_STATUS
	VIEW_MOUNTPOINTS
)

const (
	CURRENT_DIRECTORY = "Current directory: %s"
	CONFIRM_DELETE    = "Are you sure want to delete the %s?"
	FILE_NAME_MSG     = "Enter new file name: %s"
)

type MainInterface struct {
	files      []table.Row
	directory  string
	textBar    string
	table      table.Model
	fNameInput textinput.Model
	uiStatus   uint8
	fileData   viewport.Model
	style      lipgloss.Style
	ts         table.Styles
	keys       hotKeys
	helpM      help.Model
}

func (m MainInterface) Init() tea.Cmd {
	return textinput.Blink
}

func (m MainInterface) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		hexData  string
		cmd      tea.Cmd
		height   int = 4
		helpCoef int = 1
	)

	switch ms := msg.(type) {
	case tea.WindowSizeMsg:
		m.style.Width(ms.Width - 4)
		m.style.Height(ms.Height - (height + helpCoef))
		m.helpM.Width = ms.Width

		switch m.uiStatus {
		case NORMAL_STATUS:
			m.table.SetWidth(ms.Width - 4)
			m.table.SetHeight(ms.Height - (height + helpCoef))
			m.table.UpdateViewport()
		case READ_FILE:
			m.fileData.Width = ms.Width - 4
			m.fileData.Height = ms.Height - (height + helpCoef)
		}
	case tea.KeyMsg:
		switch ms.String() {
		case "h":
			m.helpM.ShowAll = !m.helpM.ShowAll
			m.table.SetHeight(lipgloss.Height(m.View()) + (height + helpCoef))
			m.table.UpdateViewport()
			m.showFiles()
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
				m.directory = filepath.Join(m.directory, m.table.SelectedRow()[2])
				m.textBar = CONFIRM_DELETE
				m.table.Blur()
				return m, nil
			}
		case tea.KeyEsc.String(), tea.KeyEscape.String():
			if m.uiStatus == DELETE_STATUS {
				m.uiStatus = NORMAL_STATUS
				m.directory = filepath.Dir(m.directory)
				m.textBar = CURRENT_DIRECTORY
				m.table.Focus()
			}

			if m.uiStatus == READ_FILE || m.uiStatus == DELETE_STATUS ||
				m.uiStatus == RENAME_FILE || m.uiStatus == CREATE_STATUS {
				m.table.Focus()
				m.uiStatus = NORMAL_STATUS
			}

			if m.uiStatus == VIEW_MOUNTPOINTS {
				m.uiStatus = NORMAL_STATUS
				m.showFiles()
			}

			return m, nil
		case tea.KeyCtrlC.String():
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
			if m.table.Focused() && m.uiStatus == NORMAL_STATUS {
				m.directory = filepath.Dir(m.directory)
				m.files = updateDirectory(m.directory)
				m.table.SetRows(updateDirectory(m.directory))
				m.table.SetCursor(0)
			}
		case tea.KeyEnter.String(), "return":
			switch m.uiStatus {
			case NORMAL_STATUS:
				m.directory = filepath.Join(m.directory, m.table.SelectedRow()[2])

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
			case RENAME_FILE:
				os.Rename(
					m.directory,
					filepath.Join(
						filepath.Dir(m.directory),
						m.fNameInput.Value(),
					),
				)
				m.fNameInput.Blur()
				m.table.Focus()
				m.uiStatus = NORMAL_STATUS
				m.directory = filepath.Dir(
					m.directory,
				)
				m.files = updateDirectory(m.directory)
				m.table.SetRows(m.files)
			case CREATE_STATUS:
				os.Chdir(m.directory)
				newDir := m.fNameInput.Value()
				err := os.Mkdir(
					newDir,
					0755,
				)
				if err != nil {
					panic(err)
				}
				m.fNameInput.Blur()
				m.table.Focus()
				m.files = updateDirectory(m.directory)
				m.table.SetRows(m.files)
				m.uiStatus = NORMAL_STATUS
			case DELETE_STATUS:
				os.Chdir(filepath.Dir(m.directory))
				os.RemoveAll(filepath.Base(m.directory))
				m.table.Focus()
				m.directory = filepath.Dir(m.directory)
				m.files = updateDirectory(m.directory)
				m.table.SetRows(m.files)
				m.uiStatus = NORMAL_STATUS
			case VIEW_MOUNTPOINTS:
				m.directory = m.table.SelectedRow()[0]
				os.Chdir(m.directory)
				m.table.Focus()
				m.files = updateDirectory(m.directory)
				m.showFiles()
				m.uiStatus = NORMAL_STATUS
			}
		case tea.KeyCtrlR.String():
			if m.uiStatus == NORMAL_STATUS {
				m.uiStatus = RENAME_FILE
				m.table.Blur()
				m.directory = filepath.Join(
					m.directory, m.table.SelectedRow()[2],
				)
				m.fNameInput.Focus()
			}
		case tea.KeyCtrlT.String():
			if m.uiStatus == NORMAL_STATUS {
				m.uiStatus = CREATE_STATUS
				m.table.Blur()
				m.fNameInput.Focus()
			}
		case tea.KeyCtrlD.String():
			if m.uiStatus == NORMAL_STATUS {
				m.uiStatus = VIEW_MOUNTPOINTS
				m.showMountPoints()
				m.table.Focus()
			}
		}
	}

	switch m.helpM.ShowAll {
	case true:
		helpCoef = 4
	case false:
		helpCoef = 1
	}

	m.fNameInput, cmd = m.fNameInput.Update(msg)

	return m, cmd
}

func (m *MainInterface) showMountPoints() {
	rows := []table.Row{}

	mountPoints, _ := mounts.GetMounts()
	for _, i := range mountPoints {
		rows = append(rows,
			table.Row{
				i,
			})
	}

	m.table = table.New(
		table.WithColumns(
			[]table.Column{
				{
					Title: "Mount Points",
					Width: func() int {
						w, _, _ := term.GetSize(0)
						return w
					}(),
				},
			}),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithStyles(m.ts),
	)
}

func (m *MainInterface) showFiles() {
	w, _, _ := term.GetSize(0)
	m.table = table.New(
		table.WithColumns(
			[]table.Column{
				{
					Title: "Permissions",
					Width: w / 6,
				},
				{
					Title: "Owner",
					Width: w / 6,
				},
				{
					Title: "File name",
					Width: w / 6,
				},
				{
					Title: "File size",
					Width: w / 6,
				},
				{
					Title: "Last Modified",
					Width: w / 6,
				},
			},
		),
		table.WithRows(updateDirectory(m.directory)),
		table.WithFocused(true),
		table.WithStyles(m.ts),
	)
}

func (m MainInterface) View() string {
	return m.style.Render(
		func() string {
			if m.uiStatus == NORMAL_STATUS ||
				m.uiStatus == DELETE_STATUS ||
				m.uiStatus == VIEW_MOUNTPOINTS {
				return lipgloss.JoinVertical(
					lipgloss.Top,
					m.table.View(),
					fmt.Sprintf(
						m.textBar,
						m.directory),
					m.helpM.View(m.keys),
				)
			} else if m.uiStatus == RENAME_FILE ||
				m.uiStatus == CREATE_STATUS {
				return fmt.Sprintf(
					FILE_NAME_MSG, m.fNameInput.View(),
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

	upperF, _ := os.Stat("..")
	if directory != "/" {
		rows = append(rows, table.Row{
			upperF.Mode().String(),
			GetFileOwner(upperF),
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

		var (
			size        float32 = 0
			measurement int     = 0
		)

		rows = append(
			rows, table.Row{
				stat.Mode().String(),
				GetFileOwner(stat),
				i.Name(),
				func() string {
					if !stat.IsDir() {
						return fmt.Sprintf("%.2f %s",
							func() float32 {
								size, measurement = getMBSize(
									float32(stat.Size()),
								)
								return size
							}(),
							func() string {
								switch measurement {
								case 0:
									return "MB"
								case 1:
									return "KB"
								default:
									return "Undefined"
								}
							}(),
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

func InitialMode() MainInterface {
	var m = MainInterface{
		keys:  keys,
		helpM: help.New(),
	}
	var err error
	w, h, _ := term.GetSize(0)

	m.uiStatus = NORMAL_STATUS
	m.textBar = CURRENT_DIRECTORY
	m.fileData = viewport.New(w, h)
	m.fNameInput = textinput.New()

	if m.directory, err = os.Getwd(); err != nil {
		panic(err)
	}

	m.showFiles()

	m.style = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Top).
		AlignVertical(lipgloss.Top).
		Border(lipgloss.DoubleBorder(), true).
		Width(w)
	m.ts = table.DefaultStyles()
	m.ts.Header = m.ts.Header.
		Foreground(lipgloss.Color("33"))
	m.ts.Selected = m.ts.Selected.
		Foreground(lipgloss.Color("201")).
		Bold(true)
	m.table.SetStyles(m.ts)

	helpBaseStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("33"))
	m.helpM.Styles = help.New().Styles
	m.helpM.Styles.FullDesc = helpBaseStyle
	m.helpM.Styles.ShortDesc = helpBaseStyle
	m.helpM.Styles.FullKey = helpBaseStyle.Foreground(lipgloss.Color("2"))
	m.helpM.Styles.ShortKey = helpBaseStyle.Foreground(lipgloss.Color("2"))
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
