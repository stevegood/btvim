package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

type Model struct {
	currentPath string
	textarea    textarea.Model
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			// TODO: prompt user to see if they want to quit?
			// should we even do this since we want to support :q :wq :q!, etc later?
			return m, tea.Quit
		case "esc":
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.textarea.SetWidth(msg.Width - h)
		m.textarea.SetHeight(msg.Height - v)
		if !m.isDir() {
			m.textarea.SetValue(m.fileContent())
		}
	}

	m.textarea, cmd = m.textarea.Update(message)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder

	switch m.isDir() {
	case true:
		b.WriteString("tree...")
	case false:
		b.WriteString(m.textarea.View())
	}

	return b.String()
}

func (m Model) isDir() bool {
	tail := string(m.currentPath[len(m.currentPath)-1])
	return tail == "/" || tail == "."
}

func (m Model) fileContent() string {
	b, err := os.ReadFile(m.currentPath)
	if err == os.ErrNotExist {
		return ""
	}

	return string(b)
}

func NewModel(currentPath string) Model {
	ti := textarea.New()
	ti.Focus()

	return Model{
		currentPath: currentPath,
		textarea:    ti,
	}
}
