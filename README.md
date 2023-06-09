Vim Bubble
==========

This repository contains a [BubbleTea](https://github.com/charmbracelet/bubbletea) [Bubble](https://github.com/charmbracelet/bubbles) that emulates some common Vim functionality inside of the BubbleTea framework. It is built on a heavily-modified fork of Charm's [textarea Bubble](https://github.com/charmbracelet/bubbles#text-area).

Loom demo video: https://www.loom.com/share/40b389a2e694408aa2b74d1f9b02d1dc

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
