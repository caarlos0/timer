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
man timer
timer --help
```

It is possible to pass a time unit for `<duration>`.

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
If no unit is passed, it defaults to seconds ("s").

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

