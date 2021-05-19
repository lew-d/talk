package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	lorem "github.com/drhodes/golorem"
)

var (
	ready = false

	width          = 0
	height         = 0
	nameWidth      = 10
	seperatorWidth = 3

	lightGrey     = lipgloss.NewStyle().Foreground(lipgloss.Color("#555"))
	nameStyle     = lipgloss.NewStyle().Width(nameWidth).MaxWidth(nameWidth).Align(lipgloss.Right).MaxHeight(10)
	nameSeperator = lipgloss.NewStyle().Inherit(lightGrey).Render(" : ")
	contentsStyle = lipgloss.NewStyle()
	inputPad      = strings.Repeat(" ", nameWidth+1) + nameSeperator
)

func AsStringSlice(in []fmt.Stringer) []string {
	out := []string{""}
	for _, i := range in {
		out = append(out, i.String())
	}
	return out
}

type incomingMessage struct {
	mode     string
	user     string
	contents string
}

type buffer []fmt.Stringer

type model struct {
	buffer   buffer
	input    textinput.Model
	viewport viewport.Model
}

func (m incomingMessage) String() string {
	return lipgloss.JoinHorizontal(0,
		m.mode,
		nameStyle.Render(m.user),
		nameSeperator,
		contentsStyle.Width(width-nameWidth-seperatorWidth-1).Render(m.contents),
	)
}

func NewModel() model {
	ti := textinput.NewModel()
	ti.Focus()
	ti.Prompt = ""
	return model{
		input: ti,
	}
}

func main() {
	p := tea.NewProgram(NewModel())
	p.EnterAltScreen()
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func NewMessage() tea.Msg {
	long := lorem.Paragraph(1, 2)
	return incomingMessage{"+", "ldewiowefwdew", long}
}

func (m model) Init() tea.Cmd {
	var messages []tea.Cmd
	for i := 1; i < 100; i++ {
		messages = append(messages, NewMessage)
	}
	return tea.Batch(append(messages, textinput.Blink)...)
}

func bufferToViewport(b buffer) (string, int) {
	bufferAsStringSlice := AsStringSlice(b)
	joined := lipgloss.JoinVertical(0, bufferAsStringSlice...)

	return joined, lipgloss.Height(joined)
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var toRender string

	switch msg := message.(type) {

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		} else {
			m.input, cmd = m.input.Update(message)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		width, height = msg.Width, msg.Height-1

		if !ready {
			m.viewport = viewport.Model{Width: width, Height: height}
		}

		m.viewport.Width, m.viewport.Height = width, height

		toRender, m.viewport.YOffset = bufferToViewport(m.buffer)
		m.viewport.SetContent(toRender)
	case incomingMessage:
		m.buffer = append(m.buffer, msg)
		toRender, m.viewport.YOffset = bufferToViewport(m.buffer)
		m.viewport.SetContent(toRender)
	}

	return m, nil
}

func (m model) View() string {
	input := inputPad + m.input.View()

	return lipgloss.JoinVertical(0, m.viewport.View(), input)
}
