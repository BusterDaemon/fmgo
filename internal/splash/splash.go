package splash

import (
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type SplashScreen struct {
	mainScreen viewport.Model
	style      lipgloss.Style
}

func (s SplashScreen) Init() tea.Cmd {
	return nil
}

func (s SplashScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch ms := msg.(type) {
	case tea.KeyMsg:
		switch ms.String() {
		case tea.KeyEnter.String():
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s SplashScreen) View() string {
	return s.style.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			s.mainScreen.View(),
		),
	)
}

func New() SplashScreen {
	return SplashScreen{
		mainScreen: func() viewport.Model {
			var (
				width  int
				height int
			)
			width, height, err := term.GetSize(int(os.Stdin.Fd()))
			if err != nil {
				width = 64
				height = 64
			}
			content := viewport.New(width-2, height-2)
			content.SetContent(
				`
				Программа "Файловый менеджер"
				Выполнена студентом группы 23ВВВ3 Точновым Е.Е.
				Нажмите Enter, чтобы продолжить
				`,
			)

			return content
		}(),
		style: lipgloss.NewStyle().
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Border(lipgloss.BlockBorder()),
	}
}
