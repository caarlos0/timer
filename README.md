<p align="center">
	<img src="https://vhs.charm.sh/vhs-1q9uo4hkLStOL4UHZFfQ4W.gif" alt="Made with VHS">
	<a href="https://vhs.charm.sh">
		<img src="https://stuff.charm.sh/vhs/badge.svg">
	</a>
	<br>
	<h1 align="center">timer</h1>
	<p align="center">A <code>sleep</code> with progress.</p>
</p>

---

Timer is a small CLI, similar to the `sleep` everyone already knows and love,
with a couple of extra features:

- a progress bar indicating the progression of said timer
- a timer showing how much time is left
- named timers

## Usage

```sh
timer <duration>
timer -n <name> <duration>
timer -t <time>
timer -t <time> -n <name>
man timer
timer --help
```

You can use the timer in two ways:

1. **Duration-based timer** (original behavior): Specify how long to wait
2. **Time-based timer** (new feature): Specify when to stop waiting

### Duration-based timer

It is possible to pass a time unit for `<duration>`.

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
If no unit is passed, it defaults to seconds ("s").

### Time-based timer

Use the `-t` or `--time` flag to specify a target time. The timer will count down until that time is reached.

Supported time formats:
- `14:30` - 24-hour format
- `2:30PM` or `2:30pm` - 12-hour format with AM/PM
- `14:30:45` - 24-hour format with seconds
- `2:30:45PM` or `2:30:45pm` - 12-hour format with seconds and AM/PM

Examples:
```sh
timer -t 14:30        # Timer until 2:30 PM today (or tomorrow if it's already past 2:30 PM)
timer -t 2:30PM       # Same as above, but using 12-hour format
timer -t 02:14am      # Timer until 2:14 AM (tomorrow morning)
timer -t 23:59:59     # Timer until 11:59:59 PM
```

If the specified time is in the past (earlier today), the timer will automatically target that time tomorrow.

If you want to show the start time in 24-hour format, use `--format 24h`. For
example:
```sh
timer 5s --format 24h -n Demo
timer -t 14:30 --format 24h -n "Until 2:30 PM"
```
Currently, the two formats supported by the `--format` option are:
- `kitchen`: the default, example: `9:16PM`.
- `24h`: 24-hour time format, example: `21:16`.

## Install

**homebrew**:

```sh
brew install caarlos0/tap/timer
```

**macports**:

```sh
sudo port install timer
```

**snap**:

```sh
snap install timer
```

**apt**:

```sh
echo 'deb [trusted=yes] https://repo.caarlos0.dev/apt/ /' | sudo tee /etc/apt/sources.list.d/caarlos0.list
sudo apt update
sudo apt install timer
```

**yum**:

```sh
echo '[caarlos0]
name=caarlos0
baseurl=https://repo.caarlos0.dev/yum/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/caarlos0.repo
sudo yum install timer
```

**arch linux**:

```sh
yay -S timer-bin
```

**deb/rpm/apk**:

Download the `.apk`, `.deb` or `.rpm` from the [releases page][releases] and install with the appropriate commands.

**manually**:

Download the pre-compiled binaries from the [releases page][releases] or clone the repo build from source.

[releases]:  https://github.com/caarlos0/timer/releases

# Badges

[![Release](https://img.shields.io/github/release/caarlos0/timer.svg?style=for-the-badge)](https://github.com/caarlos0/timer/releases/latest)

[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](LICENSE.md)

[![Build](https://img.shields.io/github/actions/workflow/status/caarlos0/timer/build.yml?style=for-the-badge)](https://github.com/caarlos0/timer/actions?query=workflow%3Abuild)

[![Go Report Card](https://goreportcard.com/badge/github.com/caarlos0/timer?style=for-the-badge)](https://goreportcard.com/report/github.com/caarlos0/timer)

[![Powered By: GoReleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=for-the-badge)](https://github.com/goreleaser)
