`3mux` is a terminal multiplexer with out-of-the-box support for search, mouse-controlled scrollback, and i3-like keybindings. Imagine `tmux` with a smaller learning curve and more user-friendly defaults.

[<img src="./demo.gif" width="800"/>](https://streamable.com/m2r57p)

<!--TODO: GIF!-->

### Features

* batteries included
* i3-like keybindings
* session management
  * optionally interactive
  * self-documenting
* search
* scrollback
* mouse support
  * drag to resize panes
  * click to select pane
  * scrollwheel

### Key Bindings

| Key(s) | Description
|-------:|:------------
|<kbd>Alt+Enter</kbd><br><kbd>Alt+N</kbd> | Create a new pane
|<kbd>Alt+Shift+F</kbd> | Make the selected pane fullscreen. Useful for copying text
|<kbd>Alt+&larr;/&darr;/&uarr;/&rarr;</kbd><br><kbd>Alt+h/j/k/l</kbd> | Select an adjacent pane
|<kbd>Alt+Shift+&larr;/&darr;/&uarr;/&rarr;</kbd><br><kbd>Alt+Shift+h/j/k/l</kbd> | Move the selected pane
|<kbd>Alt+R</kbd> | Enter resize mode. Resize selected pane with arrow keys or <kbd>h/j/k/l</kbd>. Exit using any other key(s)
|<kbd>Alt+/</kbd> | Enter search mode. Type query, navigate between results with arrow keys or <kbd>n/N</kbd>
|<kbd>Scroll</kbd> | Move through scrollback
|<kbd>Shift</kbd> | Many terminal emulators support selecting text while pressing this key


### Supported tmux Bindings

| Key(s) | Description
|-------:|:------------
|<kbd>Ctrl+b "</kbd> | Split horizontally
|<kbd>Ctrl+b %</kbd> | Split vertically
|<kbd>Ctrl+b {</kbd> | Move pane left
|<kbd>Ctrl+b }</kbd> | Move pane right

### Supported screen Bindings

| Key(s) | Description
|-------:|:------------
|<kbd>Ctrl+a \|</kbd> | Split horizontally
|<kbd>Ctrl+a S</kbd> | Split vertically
|<kbd>Ctrl+a Tab</kbd> | Cycle forward through panes

### Installation Instructions

###### Using Homebrew

```
brew update
brew install 3mux
```

###### Using Nix flakes (requires Nix 2.4+)

```
nix run github:aaronjanse/3mux
```

###### Package manager

[![Packages for 3mux](https://repology.org/badge/vertical-allrepos/3mux.svg)](https://repology.org/project/3mux/versions)

###### Building from source
1. Install Golang
2. `go get github.com/aaronjanse/3mux`
3. Run `3mux` to launch the terminal multiplexer

To update `3mux`, run `go get -u github.com/aaronjanse/3mux`

#### Terminal.app
_**Warning: Arrow-key-controlled pane management is currently unsupported on Terminal.app. Please use the default vim-like keybindings instead.**_  
Preferences > Profiles > Keyboard > Use Option as Meta Key  

#### iTerm2
Preferences > Profiles > Keys > Option Key > Esc+

### Miscellaneous

3mux searches `XDG_CONFIG_HOME` to find its config. If it cannot, it writes a config to `~/.config/3mux/config.toml` upon the first run. Modifiers in shortcuts (e.g. `Alt`) are case-insensitive.

You can detect if you're running a script inside 3mux by checking if `THREEMUX` is equal to `1`.

### Contributing
All help is welcome! You can help the project by filing issues recording what works well, what doesn't work well, and/or a feature you want. Pull Requests would be very much appreciated.

### Related Projects
* [tmux-tilish](https://github.com/jabirali/tmux-tilish)
