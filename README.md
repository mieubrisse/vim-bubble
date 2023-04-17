Vim Bubble
==========

This repository contains a [BubbleTea](https://github.com/charmbracelet/bubbletea) [Bubble](https://github.com/charmbracelet/bubbles) that emulates some common Vim functionality inside of the BubbleTea framework. It is built on a heavily-modified fork of Charm's [textarea Bubble](https://github.com/charmbracelet/bubbles#text-area).

How to use
----------
```go
```

Functionality
-------------
### Supported
- Normal & insert modes (via `i` and `a`)
- Common movement commands (h/j/k/l, `w`, `e`, `b`, `ge`, `^`, `$`, `0`, `gg`, `G`)
- Common editing functionality (`dd`, `cc`, `D`, `C`, `x`, `p`, etc.)
- Undo & redo

### Not supported but probably will
- Numbered-repeat (e.g. `20j`)
- `t`,`f`,`;`,`,`
- Visual mode
- Page up/down

### Not supported & not planning (PRs welcome)
- Registers
- Little vs big word distinction
- `%` for jump-to-matching support
- `zz`

Why?
----
I needed a text area for a BubbleTea app I was building, and I wasn't happy with the limited functionality of the default textarea. 
