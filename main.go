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
	ready     = false
	collapsed = false

	width          = 0
	height         = 0
	nameWidth      = 5
	seperatorWidth = 3
	treeWidth      = 10

	lightGreyColour = lipgloss.Color("#555")
	lightGrey       = lipgloss.NewStyle().Foreground(lightGreyColour)
	nameStyle       = lipgloss.NewStyle().Width(nameWidth).MaxWidth(nameWidth).Align(lipgloss.Right)
	treeStyle       = lipgloss.NewStyle().PaddingRight(1).Width(treeWidth).MaxWidth(treeWidth)
	barTreeStyle    = lipgloss.NewStyle().PaddingRight(1).Inherit(lightGrey)
	channelStyle    = lipgloss.NewStyle().BorderLeft(true).BorderStyle(lipgloss.NormalBorder()).PaddingLeft(1).BorderForeground(lightGreyColour)
	seperator       = "" //lipgloss.NewStyle().Inherit(lightGrey).Render(" : ") we do this somewhere else now
	contentsStyle   = lipgloss.NewStyle()
	inputPad        = strings.Repeat(" ", nameWidth+1) + seperator
)

func AsStringSlice(in []fmt.Stringer) []string {
	out := []string{}
	for _, i := range in {
		out = append(out, i.String())
	}
	return out
}

type server struct {
	name     string
	channels []string
}

type tree []server

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
	tree     tree
}

func (m incomingMessage) String() string {
	return lipgloss.JoinHorizontal(0,
		m.mode,
		nameStyle.Render(m.user),
		seperator,
		contentsStyle.Width(width-nameWidth-seperatorWidth-1).Render(m.contents),
	)
}

func (t tree) String() string {
	tr := lightGrey.Render("use/talk")
	for _, s := range t {
		tr = lipgloss.JoinVertical(0, tr, s.name)
		for _, c := range s.channels {
			tr = lipgloss.JoinVertical(0, tr, channelStyle.Render(c))
		}
	}

	if !collapsed {
		return treeStyle.Render(tr)
	} else {
		return ""
	}
}

func (t tree) StringAsBar() string {
	tr := lightGrey.Render("")
	for _, s := range t {
		for _, c := range s.channels {
			tr = lipgloss.JoinHorizontal(0, tr, barTreeStyle.Render(c))
		}
	}

	if collapsed {
		return tr
	} else {
		return ""
	}
}

func NewModel() model {
	ti := textinput.NewModel()
	ti.Focus()
	ti.Prompt = ""
	return model{
		input: ti,
		tree: []server{
			{channels: []string{
				"#rice", "#code", "#riz",
			}, name: "freenode"},
		},
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
	return incomingMessage{"+", "lew", long}
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

func (m model) HandleKey(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, nil
		default:
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var toRender string

	switch msg := message.(type) {

	case tea.KeyMsg:
		return m.HandleKey(msg)
	case tea.WindowSizeMsg:
		width, height = msg.Width-treeWidth, msg.Height-1

		//logic for responsive design
		//ignore
		if width < 100 {
			collapsed = true
			seperator = " "
			width += treeWidth

			if height > 15 {
				height--
			}
		} else {
			collapsed = false
			seperator = lipgloss.NewStyle().Inherit(lightGrey).Render(" : ")
		}

		if !ready {
			m.viewport = viewport.Model{Width: width, Height: height}
		}

		inputPad = strings.Repeat(" ", nameWidth+1) + seperator

		m.viewport.Width, m.viewport.Height = width, height

		toRender, m.viewport.YOffset = bufferToViewport(m.buffer)
		m.viewport.SetContent(toRender)

		return m, cmd
	case incomingMessage:
		m.buffer = append(m.buffer, msg)
		toRender, m.viewport.YOffset = bufferToViewport(m.buffer)
		m.viewport.SetContent(toRender)

	}

	return m, nil
}

func (m model) View() string {
	input := inputPad + m.input.View()

	return lipgloss.JoinHorizontal(
		0,
		m.tree.String(),
		lipgloss.JoinVertical(
			0,
			m.tree.StringAsBar(),
			m.viewport.View(),
			input,
		),
	)

}
