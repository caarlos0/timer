package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	mcobra "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/spf13/cobra"
)

type model struct {
	name         string
	duration     time.Duration
	start        time.Time
	timer        timer.Model
	progress     progress.Model
	quitting     bool
	interrupting bool
}

func (m model) Init() tea.Cmd {
	return m.timer.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmds []tea.Cmd
		var cmd tea.Cmd

		step := 100.0 / (m.duration).Seconds()
		cmds = append(cmds, m.progress.IncrPercent(step/100.0))

		m.timer, cmd = m.timer.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		termWidth, termHeight = msg.Width, msg.Height
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
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
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
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
		return "\n"
	}

	result := boldStyle.Render(m.start.Format(time.Kitchen))
	if m.name != "" {
		result += ": " + italicStyle.Render(m.name)
	}
	result += " - " + boldStyle.Render(m.timer.View()) + "\n" + m.progress.View()
	if *altscreen {
		textWidth, textHeight := lipgloss.Size(result)
		marginWidth := (termWidth - textWidth) / 2
		marginHeight := (termHeight - textHeight) / 2
		return lipgloss.NewStyle().Margin(marginHeight, marginWidth).Render(result)
	}
	return result
}

var (
	name        string
	altscreen   *bool
	termWidth   int
	termHeight  int
	version     = "dev"
	quitKeys    = key.NewBinding(key.WithKeys("esc", "q"))
	intKeys     = key.NewBinding(key.WithKeys("ctrl+c"))
	boldStyle   = lipgloss.NewStyle().Bold(true)
	italicStyle = lipgloss.NewStyle().Italic(true)
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
		var (
			initModel tea.Model = model{
				duration: duration,
				timer:    timer.NewWithInterval(duration, time.Second),
				progress: progress.New(progress.WithDefaultGradient()),
				name:     name,
				start:    time.Now(),
			}
			m tea.Model
		)
		if *altscreen {
			m, err = tea.NewProgram(initModel, tea.WithAltScreen()).Run()
		} else {
			m, err = tea.NewProgram(initModel).Run()
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
	RunE: func(cmd *cobra.Command, args []string) error {
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
	altscreen = rootCmd.Flags().BoolP("alt-screen", "a", false, "start in alt-screen")

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
