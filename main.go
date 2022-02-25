package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/caarlos0/go-shellwords"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/coral"
	mcoral "github.com/muesli/mango-coral"
	"github.com/muesli/roff"
)

type model struct {
	pomodoros []pomodoro
	current   int
	loops     int
	runs      []string
	start     time.Time
	timer     timer.Model
	progress  progress.Model
	quitting  bool
}

func (m model) Init() tea.Cmd {
	return m.timer.Init()
}

type nextPomodoroMsg struct {
	seq, loops int
}

func (m model) notify() error {
	for _, cmd := range m.runs {
		var b bytes.Buffer
		if err := template.
			Must(template.New("cmd").Parse(cmd)).
			Execute(&b, m.pomodoros[m.current]); err != nil {
			return fmt.Errorf("failed to parse run: %q", cmd)
		}
		parts, err := shellwords.Parse(b.String())
		if err != nil {
			return fmt.Errorf("failed to parse run: %q", cmd)
		}
		cmd := parts[0]
		var remainder []string
		if len(parts) > 1 {
			remainder = append(remainder, parts[1:]...)
		}
		c := exec.Command(cmd, remainder...)
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		if err := c.Run(); err != nil {
			return fmt.Errorf("failed run: %q: %w", cmd, err)
		}
	}

	return nil
}

func (m model) nextPomodoroCmd() func() tea.Msg {
	return func() tea.Msg {
		if err := m.notify(); err != nil {
			panic(err) // TODO: handle this properly
		}
		i := m.current + 1
		loops := m.loops
		if i >= len(m.pomodoros) {
			i = 0
			if loops != -1 {
				loops -= 1
			}
		}
		return nextPomodoroMsg{
			seq:   i,
			loops: loops,
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmds []tea.Cmd

		d := m.pomodoros[m.current].Duration.Seconds()
		step := 100.0 / d
		pct := step * (d - m.timer.Timeout.Seconds())
		cmds = append(cmds, m.progress.SetPercent(pct/100.0))

		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
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
		return m, m.nextPomodoroCmd()

	case nextPomodoroMsg:
		m.current = msg.seq
		m.loops = msg.loops
		if m.loops == 0 {
			m.quitting = true
			return m, tea.Quit
		}
		m.timer = m.pomodoros[m.current].newTimer()
		return m, m.timer.Init()

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

func (m model) View() string {
	if m.quitting {
		return "all done!\n"
	}

	result := boldStyle.Render(m.start.Format(time.Kitchen))
	result += ": " + italicStyle.Render(m.pomodoros[m.current].Name)
	result += " - " + boldStyle.Render(m.timer.View()) + "\n" + m.progress.View()
	return result
}

type pomodoro struct {
	Name     string
	Duration time.Duration
}

var (
	version  = "dev"
	quitKeys = key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	)
	boldStyle   = lipgloss.NewStyle().Bold(true)
	italicStyle = lipgloss.NewStyle().Italic(true)
)

const (
	padding  = 2
	maxWidth = 80
)

func (p pomodoro) newTimer() timer.Model {
	return timer.NewWithInterval(p.Duration, time.Second)
}

func init() {
	rootCmd.AddCommand(manCmd)
	rootCmd.Flags().IntVarP(&loops, "loops", "l", 1, "how many times should we loop the given timers")
	rootCmd.Flags().StringSliceVarP(&runs, "run", "r", []string{}, "commands to run when a timer finishes")
	rootCmd.Flags().StringVarP(&name, "name", "n", "unnamed", "timer name. applied only to unnamed timers")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var (
	name  string
	runs  []string
	loops int

	rootCmd = &coral.Command{
		Use:   "timer",
		Short: "timer is like sleep, but with progress report",
		Long: `
You can start a simple timer with:

	timer work=50m

Or simply:

	timer 10m

You can also set multiple timers, and loop between them, for instance:

	timer --loops 10 work=50m rest=10m

Finally, you can execute multiple commands when a timer finishes:

	timer \
		--loops 10 \
		--run "tput bell" \
		--run "say '{{.Name}} is done!" \
		--run "terminal-notifier -message '{{ .Name }} is done'" \
		work=25m rest=5m

Pretty much a pomodoro timer on your terminal :)

All these options should give you a quite a bit of possibilities.
		`,
		Version:      version,
		SilenceUsage: true,
		Args:         coral.ArbitraryArgs,
		RunE: func(cmd *coral.Command, args []string) error {
			var pomodoros []pomodoro
			for _, arg := range args {
				parts := strings.SplitN(arg, "=", 2)

				if len(parts) == 1 {
					parts = []string{name, parts[0]}
				}

				if len(parts) != 2 {
					return fmt.Errorf("invalid arg: %q", arg)
				}
				d, err := time.ParseDuration(parts[1])
				if err != nil {
					return fmt.Errorf("invalid arg: %q: %w", arg, err)
				}

				pomodoros = append(pomodoros, pomodoro{
					Name:     parts[0],
					Duration: d,
				})
			}

			if len(pomodoros) == 0 {
				return fmt.Errorf("need to pass the timers in the format name=duration, for example, pomodoro work=25m rest=5m")
			}

			first := pomodoros[0]
			m := model{
				timer:     first.newTimer(),
				current:   0,
				progress:  progress.New(progress.WithDefaultGradient()),
				pomodoros: pomodoros,
				start:     time.Now(),
				loops:     loops,
				runs:      runs,
			}
			return tea.NewProgram(m).Start()
		},
	}
	manCmd = &coral.Command{
		Use:                   "man",
		Short:                 "Generates man pages",
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Args:                  coral.NoArgs,
		RunE: func(cmd *coral.Command, args []string) error {
			manPage, err := mcoral.NewManPage(1, rootCmd)
			if err != nil {
				return err
			}

			_, err = fmt.Fprint(os.Stdout, manPage.Build(roff.NewDocument()))
			return err
		},
	}
)
