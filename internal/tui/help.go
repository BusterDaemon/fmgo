package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type hotKeys struct {
	Accept     key.Binding
	ParentFold key.Binding
	Cancel     key.Binding
	Delete     key.Binding
	Rename     key.Binding
	Mounts     key.Binding
	Help       key.Binding
	CreateDir  key.Binding
	Exit       key.Binding
	Arrows     key.Binding
	Home       key.Binding
	End        key.Binding
}

func (h hotKeys) ShortHelp() []key.Binding {
	return []key.Binding{
		h.Accept, h.ParentFold,
		h.Exit, h.Help,
	}
}

func (h hotKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{h.Accept, h.ParentFold, h.Cancel, h.Delete, h.Help},
		{h.Arrows, h.Home, h.End},
		{h.Rename, h.Mounts, h.CreateDir},
	}
}

var keys = hotKeys{
	Accept: key.NewBinding(
		key.WithKeys("enter", "return"),
		key.WithHelp("Enter", "Подтвердить"),
	),
	ParentFold: key.NewBinding(
		key.WithKeys(tea.KeyBackspace.String()),
		key.WithHelp("Backspace", "Род. каталог"),
	),
	Cancel: key.NewBinding(
		key.WithKeys(tea.KeyEsc.String()),
		key.WithHelp("Esc", "Отменить"),
	),
	Delete: key.NewBinding(
		key.WithKeys(tea.KeyDelete.String()),
		key.WithHelp("Delete", "Удалить"),
	),
	Rename: key.NewBinding(
		key.WithKeys(tea.KeyCtrlR.String()),
		key.WithHelp("CTRL+R", "Переименовать/Переместить"),
	),
	Mounts: key.NewBinding(
		key.WithKeys(tea.KeyCtrlD.String()),
		key.WithHelp("CTRL+D", "Диск. устройства"),
	),
	Help: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("H", "Сочетания клавиш"),
	),
	CreateDir: key.NewBinding(
		key.WithKeys(tea.KeyCtrlT.String()),
		key.WithHelp("CTRL+T", "Создать директорию"),
	),
	Exit: key.NewBinding(
		key.WithKeys(tea.KeyCtrlC.String()),
		key.WithHelp("CTRL+C", "Завершить работу"),
	),
	Arrows: key.NewBinding(
		key.WithKeys(tea.KeyUp.String()),
		key.WithHelp("↑/↓/←/→/Page Up/Page Down", "Навигация"),
	),
	Home: key.NewBinding(
		key.WithKeys(tea.KeyHome.String()),
		key.WithHelp("Home", "Переход в начало"),
	),
	End: key.NewBinding(
		key.WithKeys(tea.KeyEnd.String()),
		key.WithHelp("End", "Переход в конец"),
	),
}
