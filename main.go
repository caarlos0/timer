package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/progress"
	"github.com/charmbracelet/bubbles/v2/timer"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
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

func (m model) View() string {
	if m.quitting || m.interrupting {
		return ""
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
	result += " - " + boldStyle.Render(m.timer.View()) + "\n" + m.progress.View()
	if m.altscreen {
		return altscreenStyle.
			MarginTop((winHeight - 2) / 2).
			Render(result)
	}
	return result
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
	gradient        string
)

var gradientPresets = map[string][2]string{
	"default": {"#5A56E0", "#EE6FF8"},
	"sunset":  {"#FFB28C", "#FFC371"},
	"aqua":    {"#13547A", "#80D0C7"},
	"forest":  {"#5A3F37", "#2C7744"},
	"fire":    {"#F7971E", "#FFD200"},
}

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
		var opts []tea.ProgramOption
		if altscreen {
			opts = append(opts, tea.WithAltScreen())
		}
		interval := time.Second
		if duration < time.Minute {
			interval = 100 * time.Millisecond
		}

		// choose the gradient based on the flag
		// if the flag is empty or "default", use the default gradient
		var progressBar progress.Model
		if gradientFlag == "" || gradientFlag == "default" {
			progressBar = progress.New(progress.WithDefaultGradient())
		} else if preset, ok := gradientPresets[gradientFlag]; ok {
			progressBar = progress.New(progress.WithGradient(preset[0], preset[1]))
		} else if strings.Contains(gradientFlag, ",") {
			parts := strings.SplitN(gradientFlag, ",", 2)
			progressBar = progress.New(progress.WithGradient(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])))
		} else {
			// fallback to default gradient if invalid input
			fmt.Printf("Warning: Invalid gradient '%s'. Falling back to default gradient.\n", gradientFlag)
			progressBar = progress.New(progress.WithDefaultGradient())
		}

		m, err := tea.NewProgram(model{
			duration:        duration,
			timer:           timer.New(duration, timer.WithInterval(interval)),
			progress:        progressBar,
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
	rootCmd.Flags().StringVarP(&gradientFlag, "gradient", "g", "default", "Gradient preset (default, sunset, aqua, forest, fire) or two hex colors separated by comma (ex: --gradient=#00FF00,#0000FF)")

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
