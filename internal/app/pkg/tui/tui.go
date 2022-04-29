package tui

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fogleman/ease"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/termenv"
	configs "github.com/triumphpc/GophKeeper/internal/app/pkg/config"
	"github.com/triumphpc/GophKeeper/internal/app/service/client/grpcclient"
	"go.uber.org/zap"
	"math"
	"strconv"
	"strings"
	"time"
)

type Model struct {
	ctx           context.Context
	Choice        int
	CurrentChoice int
	Chosen        bool
	Ticks         int
	Frames        int
	Progress      float64
	Loaded        bool
	Quitting      bool

	regModel regModel
	logModel logModel
}

type logModel struct {
	focusIndex    int
	inputs        []textinput.Model
	cursorMode    textinput.CursorMode
	currentChoice int
	error         string
}

type regModel struct {
	focusIndex    int
	inputs        []textinput.Model
	cursorMode    textinput.CursorMode
	currentChoice int
	error         string
}

type tickMsg struct{}
type frameMsg struct{}

const (
	enterPage = iota
	regPage
	loginPage
	listData
	menuPage
	saveText
	saveCard
	saveFile

	progressBarWidth  = 71
	progressFullChar  = "█"
	progressEmptyChar = "░"
	tickSeconds       = 10
)

// General stuff for styling the view
var (
	term          = termenv.EnvColorProfile()
	keyword       = makeFgStyle("211")
	subtle        = makeFgStyle("241")
	progressEmpty = subtle(progressEmptyChar)
	dot           = colorFg(" • ", "236")

	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	blurredButton       = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
	helpStyle           = blurredStyle.Copy()
	errorStyle          = focusedStyle.Copy()
	focusedButton       = focusedStyle.Copy().Render("[ Submit ]")
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	// Gradient colors we'll use for the progress bar
	ramp = makeRamp("#B14FFF", "#00FFA3", progressBarWidth)
)

func InitialModel(ctx context.Context) Model {
	initialModel := Model{
		ctx,
		regPage,
		enterPage,
		false,
		tickSeconds,
		0,
		0,
		false,
		false,
		regModel{},
		logModel{},
	}

	p := tea.NewProgram(initialModel)
	if err := p.Start(); err != nil {
		configs.Instance().Logger.Fatal("could not start program:", zap.Error(err))
	}

	return initialModel
}

func initRegModel() regModel {
	m := regModel{
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Login"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}

func initLogModel() logModel {
	m := logModel{
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Login"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func frame() tea.Cmd {
	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return frameMsg{}
	})
}

// Init Main page model
func (m Model) Init() tea.Cmd {
	return tick()
}

// Update main update function.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			// Exit only on main page
			if m.CurrentChoice == enterPage {
				m.Quitting = true
				return m, tea.Quit
			}
		}
	}

	// Switch pages
	switch m.CurrentChoice {
	case enterPage:
		if !m.Chosen {
			return updateChoices(msg, m)
		}
		return updateChosen(msg, m) //todo убрать
	case regPage:
		return updateChosenRegPage(msg, &m)
	case loginPage:
		return updateChosenLoginPage(msg, &m)
	case menuPage:
		return updateMenuPage(msg, &m)
	}

	return m, nil

}

func (m *Model) toEnterPage() (tea.Model, tea.Cmd) {
	m.CurrentChoice = enterPage
	m.Ticks = 10
	m.Chosen = false

	return m, tick()
}

func (m *Model) toMenuPage() (tea.Model, tea.Cmd) {
	m.CurrentChoice = menuPage
	m.Chosen = false
	m.Choice = saveText

	return m, tick()
}

func updateChosenRegPage(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m.toEnterPage()

		// Change cursor mode
		case "ctrl+r":
			m.regModel.cursorMode++
			if m.regModel.cursorMode > textinput.CursorHide {
				m.regModel.cursorMode = textinput.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.regModel.inputs))
			for i := range m.regModel.inputs {
				cmds[i] = m.regModel.inputs[i].SetCursorMode(m.regModel.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.regModel.focusIndex == len(m.regModel.inputs) {
				var login, password string

				for _, v := range m.regModel.inputs {
					if v.Placeholder == "Login" {
						login = v.Value()
					} else if v.Placeholder == "Password" {
						password = v.Value()
					}
				}

				// try login user
				grpcClient := grpcclient.Instance()
				err := grpcClient.RegClient(m.ctx, login, password)
				if err != nil {
					m.regModel.error = err.Error()

					return m, tick()
				}
				return m.toMenuPage()
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.regModel.focusIndex--
			} else {
				m.regModel.focusIndex++
			}

			if m.regModel.focusIndex > len(m.regModel.inputs) {
				m.regModel.focusIndex = 0
			} else if m.regModel.focusIndex < 0 {
				m.regModel.focusIndex = len(m.regModel.inputs)
			}

			cmds := make([]tea.Cmd, len(m.regModel.inputs))
			for i := 0; i <= len(m.regModel.inputs)-1; i++ {
				if i == m.regModel.focusIndex {
					// Set focused state
					cmds[i] = m.regModel.inputs[i].Focus()
					m.regModel.inputs[i].PromptStyle = focusedStyle
					m.regModel.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.regModel.inputs[i].Blur()
				m.regModel.inputs[i].PromptStyle = noStyle
				m.regModel.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.regModel.updateInputs(msg)

	return m, cmd

}

func updateChosenLoginPage(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":

			return m.toEnterPage()

		// Change cursor mode
		case "ctrl+r":
			m.logModel.cursorMode++
			if m.logModel.cursorMode > textinput.CursorHide {
				m.logModel.cursorMode = textinput.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.logModel.inputs))
			for i := range m.logModel.inputs {
				cmds[i] = m.logModel.inputs[i].SetCursorMode(m.logModel.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.logModel.focusIndex == len(m.logModel.inputs) {
				var login, password string

				for _, v := range m.logModel.inputs {
					if v.Placeholder == "Login" {
						login = v.Value()
					} else if v.Placeholder == "Password" {
						password = v.Value()
					}
				}

				// try login user
				grpcClient := grpcclient.Instance()
				err := grpcClient.AuthClient(m.ctx, login, password)
				if err != nil {
					m.logModel.error = err.Error()

					return m, tick()
				}
				return m.toMenuPage()
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.logModel.focusIndex--
			} else {
				m.logModel.focusIndex++
			}

			if m.logModel.focusIndex > len(m.logModel.inputs) {
				m.logModel.focusIndex = 0
			} else if m.logModel.focusIndex < 0 {
				m.logModel.focusIndex = len(m.logModel.inputs)
			}

			cmds := make([]tea.Cmd, len(m.logModel.inputs))
			for i := 0; i <= len(m.logModel.inputs)-1; i++ {
				if i == m.logModel.focusIndex {
					// Set focused state
					cmds[i] = m.logModel.inputs[i].Focus()
					m.logModel.inputs[i].PromptStyle = focusedStyle
					m.logModel.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.logModel.inputs[i].Blur()
				m.logModel.inputs[i].PromptStyle = noStyle
				m.logModel.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.logModel.updateInputs(msg)

	return m, cmd

}

func updateMenuPage(msg tea.Msg, m *Model) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			return m.toEnterPage()
		}
	}

	return updateChoices(msg, *m)
}

// View the main view, which just calls the appropriate sub-view
func (m Model) View() string {
	var s string
	if m.Quitting {
		return "\n  See you later!\n\n"
	}

	switch m.CurrentChoice {
	case enterPage:
		if !m.Chosen {
			s = choicesView(m)
		} else {
			s = chosenView(m)
		}
		return indent.String("\n"+s+"\n\n", 2)

	case regPage:
		return m.regModel.View()

	case loginPage:
		return m.logModel.View()

	case menuPage:
		s := choicesMenuView(m)

		return indent.String("\n"+s+"\n\n", 2)
	}

	return ""
}

// View for registration
func (m regModel) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	if m.error != "" {
		b.WriteString(errorStyle.Render("\n" + m.error))
	}

	return b.String()
}

func (m logModel) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	if m.error != "" {
		b.WriteString(errorStyle.Render("\n" + m.error))
	}

	return b.String()
}

// Sub-update functions

// Update loop for the first view where you're choosing a task.
func updateChoices(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.Choice += 1
			if m.CurrentChoice == enterPage {
				if m.Choice > listData {
					m.Choice = listData
				}
			} else {
				if m.Choice > saveFile {
					m.Choice = saveCard
				}
			}
		case "k", "up":
			m.Choice -= 1
			if m.CurrentChoice == enterPage {
				if m.Choice < regPage {
					m.Choice = regPage
				}
			} else {
				if m.Choice < saveText {
					m.Choice = saveText
				}
			}
		case "enter":
			m.Chosen = true
			m.CurrentChoice = m.Choice

			if m.Choice == loginPage {
				m.logModel = initLogModel()
			} else if m.Choice == regPage {
				m.regModel = initRegModel()
			}

			return m, frame()
		}

	case tickMsg:
		if m.CurrentChoice == enterPage {
			if m.Ticks == 0 {
				m.Quitting = true
				return m, tea.Quit
			}
			m.Ticks -= 1
		}
		return m, tick()
	}

	return m, nil
}

// Update loop for the second view after a choice has been made
func updateChosen(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg.(type) {

	case frameMsg:
		if !m.Loaded {
			m.Frames += 1
			m.Progress = ease.OutBounce(float64(m.Frames) / float64(100))
			if m.Progress >= 1 {
				m.Progress = 1
				m.Loaded = true
				m.Ticks = 3
				return m, tick()
			}
			return m, frame()
		}

	case tickMsg:
		if m.Loaded {
			if m.Ticks == 0 {
				m.Quitting = true
				return m, tea.Quit
			}
			m.Ticks -= 1
			return m, tick()
		}
	}

	return m, nil
}

func (m *logModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m *regModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

// Sub-views

// The first view, where you're choosing a task
func choicesView(m Model) string {
	c := m.Choice

	tpl := "Make your chose\n\n"
	tpl += "%s\n\n"
	tpl += "Program quits in %s seconds\n\n"
	tpl += subtle("j/k, up/down: select") + dot + subtle("enter: choose") + dot + subtle("q, esc: quit")

	choices := fmt.Sprintf(
		"%s\n%s\n%s\n",
		checkbox("Registration", c == regPage),
		checkbox("Login", c == loginPage),
		checkbox("List data", c == listData),
	)

	return fmt.Sprintf(tpl, choices, colorFg(strconv.Itoa(m.Ticks), "79"))
}

// The second view, after a task has been chosen
func chosenView(m Model) string { // todo убрать
	var msg string

	switch m.Choice {
	case 0:
		//return registrationView()
		msg = fmt.Sprintf("Carrot planting?\n\nCool, we'll need %s and %s...", keyword("libgarden"), keyword("vegeutils"))

	default:
		msg = fmt.Sprintf("It’s always good to see friends.\n\nFetching %s and %s...", keyword("social-skills"), keyword("conversationutils"))
	}

	label := "Downloading..."
	if m.Loaded {
		label = fmt.Sprintf("Downloaded. Exiting in %s seconds...", colorFg(strconv.Itoa(m.Ticks), "79"))
	}

	return msg + "\n\n" + label + "\n" + progressbar(80, m.Progress) + "%"
}

// The first view, where you're choosing a task
func choicesMenuView(m Model) string {
	c := m.Choice

	tpl := "Make your chose\n\n"
	tpl += "%s\n\n"
	tpl += subtle("j/k, up/down: select") + dot + subtle("enter: choose") + dot + subtle("q, esc: quit")

	choices := fmt.Sprintf(
		"%s\n%s\n%s\n",
		checkbox("Save text", c == saveText),
		checkbox("Save card", c == saveCard),
		checkbox("Save file", c == saveFile),
	)

	return fmt.Sprintf(tpl, choices)
}

func checkbox(label string, checked bool) string {
	if checked {
		return colorFg("[x] "+label, "212")
	}
	return fmt.Sprintf("[ ] %s", label)
}

//todo убрать
func progressbar(width int, percent float64) string {
	w := float64(progressBarWidth)

	fullSize := int(math.Round(w * percent))
	var fullCells string
	for i := 0; i < fullSize; i++ {
		fullCells += termenv.String(progressFullChar).Foreground(term.Color(ramp[i])).String()
	}

	emptySize := int(w) - fullSize
	emptyCells := strings.Repeat(progressEmpty, emptySize)

	return fmt.Sprintf("%s%s %3.0f", fullCells, emptyCells, math.Round(percent*100))
}

// Utils

// Color a string's foreground with the given value.
func colorFg(val, color string) string {
	return termenv.String(val).Foreground(term.Color(color)).String()
}

// Return a function that will colorize the foreground of a given string.
func makeFgStyle(color string) func(string) string {
	return termenv.Style{}.Foreground(term.Color(color)).Styled
}

// Color a string's foreground and background with the given value.
func makeFgBgStyle(fg, bg string) func(string) string {
	return termenv.Style{}.
		Foreground(term.Color(fg)).
		Background(term.Color(bg)).
		Styled
}

// Generate a blend of colors.
func makeRamp(colorA, colorB string, steps float64) (s []string) {
	cA, _ := colorful.Hex(colorA)
	cB, _ := colorful.Hex(colorB)

	for i := 0.0; i < steps; i++ {
		c := cA.BlendLuv(cB, i/steps)
		s = append(s, colorToHex(c))
	}
	return
}

// Convert a colorful.Color to a hexadecimal format compatible with termenv.
func colorToHex(c colorful.Color) string {
	return fmt.Sprintf("#%s%s%s", colorFloatToHex(c.R), colorFloatToHex(c.G), colorFloatToHex(c.B))
}

// Helper function for converting colors to hex. Assumes a value between 0 and
// 1.
func colorFloatToHex(f float64) (s string) {
	s = strconv.FormatInt(int64(f*255), 16)
	if len(s) == 1 {
		s = "0" + s
	}
	return
}
