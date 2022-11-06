package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stevegood/btvim/pkg/editor"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

type fileSaveMsg struct {
	err error
}

type Model struct {
	currentPath      string
	err              error
	textarea         textarea.Model
	editorMode       editor.Mode
	commandModeInput textinput.Model
}

func (m Model) Init() tea.Cmd {
	return tea.Batch([]tea.Cmd{
		textarea.Blink,
		textinput.Blink,
	}...)
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch m.editorMode {
	case editor.NormalMode:
		m, cmd = m.normalModeUpdate(message)
		cmds = append(cmds, cmd)
	case editor.InsertMode:
		m, cmd = m.insertModeUpdate(message)
		cmds = append(cmds, cmd)
	case editor.CommandMode:
		m, cmd = m.commandModeUpdate(message)
		cmds = append(cmds, cmd)
	}

	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if !m.inMode(editor.NormalMode) {
				m.editorMode = editor.NormalMode
			}
		case "i":
			if m.inMode(editor.NormalMode) {
				m.editorMode = editor.InsertMode
			}
		case ":":
			if m.inMode(editor.NormalMode) {
				m.editorMode = editor.CommandMode
				m.commandModeInput.SetValue(":")
				cmd = m.commandModeInput.Focus()
				cmds = append(cmds, cmd)
			}
		default:
			if !m.textarea.Focused() && !m.inMode(editor.CommandMode) {
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

	return m, tea.Batch(cmds...)
}

func (m Model) normalModeUpdate(message tea.Msg) (Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "left", "right", "down":
			m.textarea, cmd = m.textarea.Update(message)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) insertModeUpdate(message tea.Msg) (Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)
	m.textarea, cmd = m.textarea.Update(message)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) commandModeUpdate(message tea.Msg) (Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch m.commandModeInput.Value() {
			case ":w":
				// TODO: save the file
				cmds = append(cmds, m.saveFile)
				// go back to normal mode
				m.editorMode = editor.NormalMode
			case ":q", ":q!":
				// quit
				// TODO: prompt user to save the file changes first (if there were any changes)
				return m, tea.Quit
			case ":wq":
				// TODO: save the file and then quit
				m.saveFile()
				return m, tea.Quit
			}
		}
	}

	m.commandModeInput, cmd = m.commandModeInput.Update(message)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) saveFile() tea.Msg {
	err := os.WriteFile(m.currentPath, []byte(m.textarea.Value()), os.ModePerm)
	return fileSaveMsg{
		err: err,
	}
}

func (m Model) View() string {
	var b strings.Builder

	switch m.isDir() {
	case true:
		b.WriteString("tree...")
	case false:
		b.WriteString(m.textarea.View() + "\n")
	}

	switch m.editorMode {
	case editor.NormalMode:
		b.WriteString("NORMAL")
	case editor.InsertMode:
		b.WriteString("INSERT")
	case editor.CommandMode:
		b.WriteString("COMMAND")
		b.WriteString(m.commandModeInput.View())
	}

	return b.String()
}

func (m Model) inMode(mode editor.Mode) bool {
	return m.editorMode == mode
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

	cmdInput := textinput.New()

	return Model{
		currentPath:      currentPath,
		textarea:         ti,
		editorMode:       editor.NormalMode,
		commandModeInput: cmdInput,
	}
}
