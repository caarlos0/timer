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
	endTime := m.start.Add(m.duration)

	// Format remaining time with custom precision
	remainingTime := m.duration - m.passed
	formattedTime := formatDuration(remainingTime)

	result +=
		" - " + boldStyle.Render(endTime.Format(startTimeFormat)) +
			" - " + boldStyle.Render(formattedTime) +
			"\n" + m.progress.View()
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
	targetTime      string
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
	Use:          "timer [duration]",
	Short:        "timer is like sleep, but with progress report",
	Long:         "Timer can count down from a duration or until a specific time. Use either [duration] argument or --time flag.",
	Version:      version,
	SilenceUsage: true,
	Args:         cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var duration time.Duration
		var err error

		if targetTime != "" {
			// Parse target time and calculate duration
			duration, err = calculateDurationUntilTime(targetTime)
			if err != nil {
				return fmt.Errorf("failed to parse target time: %w", err)
			}
		} else {
			// Original behavior: parse duration from args
			if len(args) != 1 {
				return fmt.Errorf("duration argument is required when --time is not specified")
			}
			addSuffixIfArgIsNumber(&(args[0]), "s")
			duration, err = time.ParseDuration(args[0])
			if err != nil {
				return err
			}
		}

		if duration <= 0 {
			return fmt.Errorf("duration must be positive")
		}

		var opts []tea.ProgramOption
		if altscreen {
			opts = append(opts, tea.WithAltScreen())
		}
		interval := time.Second
		if duration < time.Minute {
			interval = 100 * time.Millisecond
		}
		m, err := tea.NewProgram(model{
			duration:        duration,
			timer:           timer.New(duration, timer.WithInterval(interval)),
			progress:        progress.New(progress.WithDefaultGradient()),
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
	rootCmd.Flags().StringVarP(&targetTime, "time", "t", "", "timer until specific time (e.g., 14:30, 2:30PM, 02:14am)")

	rootCmd.AddCommand(manCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// addSuffixIfArgIsNumber appends a suffix to the argument if it is a number
func addSuffixIfArgIsNumber(s *string, suffix string) {
	_, err := strconv.ParseFloat(*s, 64)
	if err == nil {
		*s = *s + suffix
	}
}

// formatDuration formats a duration with clean display and 2 decimal places for seconds
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "0s"
	}
	if d == 0 {
		return "0s"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := d.Seconds() - float64(hours*3600) - float64(minutes*60)

	var parts []string

	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%.0fs", seconds))
	}

	return strings.Join(parts, "")
}

// calculateDurationUntilTime calculates the duration from now until the specified target time
func calculateDurationUntilTime(targetTimeStr string) (time.Duration, error) {
	now := time.Now()

	// Try multiple time formats
	timeFormats := []string{
		"15:04",     // 24-hour format: 14:30
		"3:04PM",    // 12-hour format with PM: 2:30PM
		"3:04pm",    // 12-hour format with pm: 2:30pm
		"15:04:05",  // 24-hour format with seconds: 14:30:45
		"3:04:05PM", // 12-hour format with seconds and PM: 2:30:45PM
		"3:04:05pm", // 12-hour format with seconds and pm: 2:30:45pm
	}

	var targetTime time.Time
	var err error

	for _, format := range timeFormats {
		if targetTime, err = time.Parse(format, targetTimeStr); err == nil {
			break
		}
	}

	if err != nil {
		return 0, fmt.Errorf("unable to parse time format. Supported formats: 15:04, 3:04PM, 3:04pm, 15:04:05, 3:04:05PM, 3:04:05pm")
	}

	// Set the target time to today
	targetTime = time.Date(now.Year(), now.Month(), now.Day(),
		targetTime.Hour(), targetTime.Minute(), targetTime.Second(), 0, now.Location())

	// Calculate duration until target time
	duration := targetTime.Sub(now)

	// Schedule for tomorrow if the time has passed or is the exact same time
	if duration <= 0 {
		targetTime = targetTime.AddDate(0, 0, 1)
		duration = targetTime.Sub(now)
	}

	return duration, nil
}
