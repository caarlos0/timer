package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/timer"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/spf13/cobra"
)

type model struct {
	name            string
	altscreen       bool
	startTimeFormat string
	duration        time.Duration
	passed          time.Duration
	start           time.Time
	timer           timer.Model
	progress        progress.Model
	quitting        bool
	interrupting    bool
}

func (m model) Init() tea.Cmd {
	return m.timer.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmds []tea.Cmd
		var cmd tea.Cmd

		m.passed += m.timer.Interval
		pct := m.passed.Milliseconds() * 100 / m.duration.Milliseconds()
		cmds = append(cmds, m.progress.SetPercent(float64(pct)/100))

		m.timer, cmd = m.timer.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.progress.SetWidth(msg.Width - padding*2 - 4)
		winHeight = msg.Height
		if !m.altscreen && m.progress.Width() > maxWidth {
			m.progress.SetWidth(maxWidth)
		}
		return m, nil

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		m.quitting = true
		return m, tea.Quit

	case progress.FrameMsg:
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}
		if key.Matches(msg, intKeys) {
			m.interrupting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	if m.quitting || m.interrupting {
		return tea.NewView("")
	}

	var startTimeFormat string
	switch strings.ToLower(m.startTimeFormat) {
	case "24h":
		startTimeFormat = "15:04" // See: https://golang.cafe/blog/golang-time-format-example.html
	default:
		startTimeFormat = time.Kitchen
	}
	result := boldStyle.Render(m.start.Format(startTimeFormat))
	if m.name != "" {
		result += ": " + italicStyle.Render(m.name)
	}
	endTime := m.start.Add(m.duration)
	result += " - " + boldStyle.Render(endTime.Format(startTimeFormat)) +
		" - " + boldStyle.Render(m.timer.View()) +
		"\n" + m.progress.View()
	if m.altscreen {
		s := altscreenStyle.
			MarginTop((winHeight - 2) / 2).
			Render(result)
		v := tea.NewView(s)
		v.AltScreen = true
		return v
	}
	return tea.NewView(result)
}

var (
	name            string
	altscreen       bool
	startTimeFormat string
	winHeight       int
	version         = "dev"
	quitKeys        = key.NewBinding(key.WithKeys("esc", "q"))
	intKeys         = key.NewBinding(key.WithKeys("ctrl+c"))
	altscreenStyle  = lipgloss.NewStyle().MarginLeft(padding)
	boldStyle       = lipgloss.NewStyle().Bold(true)
	italicStyle     = lipgloss.NewStyle().Italic(true)
)

const (
	padding  = 2
	maxWidth = 80
)

var rootCmd = &cobra.Command{
	Use:          "timer",
	Short:        "timer is like sleep, but with progress report",
	Version:      version,
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addSuffixIfArgIsNumber(&(args[0]), "s")
		duration, err := time.ParseDuration(args[0])
		if err != nil {
			return err
		}
		if duration <= 0 {
			return fmt.Errorf("timer duration cannot be set to 0 or less")
		}
		var opts []tea.ProgramOption
		interval := time.Second
		if duration < time.Minute {
			interval = 100 * time.Millisecond
		}
		m, err := tea.NewProgram(model{
			duration: duration,
			timer:    timer.New(duration, timer.WithInterval(interval)),
			progress: progress.New(progress.WithColors(
				lipgloss.Color("#5A56E0"),
				lipgloss.Color("#EE6FF8"),
			)),
			name:            name,
			altscreen:       altscreen,
			startTimeFormat: startTimeFormat,
			start:           time.Now(),
		}, opts...).Run()
		if err != nil {
			return err
		}
		if m.(model).interrupting {
			return fmt.Errorf("interrupted")
		}
		if name != "" {
			cmd.Printf("%s ", name)
		}
		cmd.Printf("finished!\n")
		return nil
	},
}

var manCmd = &cobra.Command{
	Use:                   "man",
	Short:                 "Generates man pages",
	SilenceUsage:          true,
	DisableFlagsInUseLine: true,
	Hidden:                true,
	Args:                  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		manPage, err := mcobra.NewManPage(1, rootCmd)
		if err != nil {
			return err
		}

		_, err = fmt.Fprint(os.Stdout, manPage.Build(roff.NewDocument()))
		return err
	},
}

func init() {
	rootCmd.Flags().StringVarP(&name, "name", "n", "", "timer name")
	rootCmd.Flags().BoolVarP(&altscreen, "fullscreen", "f", false, "fullscreen")
	rootCmd.Flags().StringVarP(&startTimeFormat, "format", "", "", "Specify start time format, possible values: 24h, kitchen")

	rootCmd.AddCommand(manCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func addSuffixIfArgIsNumber(s *string, suffix string) {
	_, err := strconv.ParseFloat(*s, 64)
	if err == nil {
		*s = *s + suffix
	}
}
