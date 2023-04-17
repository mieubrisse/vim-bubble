Vim Bubble
==========

This repository contains a [BubbleTea](https://github.com/charmbracelet/bubbletea) [Bubble](https://github.com/charmbracelet/bubbles) that emulates some common Vim functionality inside of the BubbleTea framework. It is built on a heavily-modified fork of Charm's [textarea Bubble](https://github.com/charmbracelet/bubbles#text-area).

<div style="position: relative; padding-bottom: 64.63195691202873%; height: 0;"><iframe src="https://www.loom.com/embed/40b389a2e694408aa2b74d1f9b02d1dc" frameborder="0" webkitallowfullscreen mozallowfullscreen allowfullscreen style="position: absolute; top: 0; left: 0; width: 100%; height: 100%;"></iframe></div>

How to use
----------
```bash
go get github.com/mieubrisse/vim-bubble
```
```go
vim := vim.New()
```

Functionality
-------------
### Supported
- Normal & insert modes (via `i` and `a`)
- Common movement commands (`h`, `j`, `k`, `l`, `w`, `e`, `b`, `ge`, `^`, `$`, `0`, `gg`, `G`, etc.)
- Common editing functionality (`dd`, `cc`, `D`, `C`, `x`, `p`, `o`, `O`, etc.)
- Undo & redo (via `u` and `ctrl+r`)

### Not supported but probably will
- GIF to demo this
- Numbered-repeat (e.g. `20j`, `10g`)
- `t`,`f`,`;`,`,`
- Visual mode
- Page up/down
- Different stylings on the UI elements

### Not supported & not planning (PRs welcome)
- Registers
- Little vs big word distinction
- `%` for jump-to-matching support
- `zz`
- `H`, `M`, `L`

Why?
----
I needed a text area for a BubbleTea app I was building, and I wasn't happy with the limited, emacs-like bindings on the default textarea. Hopefully this is useful for other Vim nerds.
