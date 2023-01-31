<p align="center">
	<img src="https://vhs.charm.sh/vhs-3JUif7gjSs5lL43g5X4LzS.gif" alt="Made with VHS">
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

## Usage

```sh
timer <duration>
```

It is possible to pass a time unit for `<duration>`.

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
If no unit is passed, it defaults to seconds ("s").

## Install

**Download**:

Download from the [releases page][rlurl]. (built on M1 mac)

[rlurl]: https://github.com/jacklolidk/just-timer/releases

**Manually**:

0. Make sure [go][gourl] is installed 
1. Clone the repo
2. Run `go build -o timer`
3. Profit

[gourl]: https://go.dev/
