package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	name     string
	start    time.Time
	timer    timer.Model
	progress progress.Model
	quitting bool
}

func (m model) Init() tea.Cmd {
	return m.timer.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmds []tea.Cmd
		var cmd tea.Cmd

		step := 100.0 / (*timerFor).Seconds()

		cmds = append(cmds, m.progress.IncrPercent(step/100.0))
		// pm, cmd := m.progress.Update(msg)
		// cmds = append(cmds, cmd)
		// m.progress = pm.(progress.Model)

		m.timer, cmd = m.timer.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		m.quitting = true
		return m, tea.Quit

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

var boldStyle = lipgloss.NewStyle().Bold(true)
var italicStyle = lipgloss.NewStyle().Italic(true)

func (m model) View() string {
	if m.quitting {
		return "\n"
	}

	return boldStyle.Render(m.start.Format(time.Kitchen)) +
		": " +
		italicStyle.Render(m.name) +
		" - " +
		boldStyle.Render(m.timer.View()) +
		" " +
		m.progress.View()
}

var timerFor = flag.Duration("for", 50*time.Minute, "how log the timer should go")
var name = flag.String("name", "", "name this timer")

func main() {
	flag.Parse()
	m := model{
		timer:    timer.NewWithInterval(*timerFor, time.Second),
		progress: progress.New(progress.WithDefaultGradient()),
		name:     *name,
		start:    time.Now(),
	}

	if err := tea.NewProgram(m).Start(); err != nil {
		fmt.Println("Uh oh, we encountered an error:", err)
		os.Exit(1)
	}
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "ctrl+c"),
	key.WithHelp("q", "quit"),
)
