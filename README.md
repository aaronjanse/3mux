# 3mux

`3mux` is a terminal multiplexer with out-of-the-box support for search, mouse-controlled scrollback, and i3-like keybindings. 

![Screenshot](./i3-tmux.png)

<!--TODO: GIF!-->

## TODO:
- [ ] finish search implementation
- [ ] support default tmux keybindings

## Key Bindings

| Key(s) | Description
|-------:|:------------
|<kbd>Alt+Enter</kbd><br><kbd>Alt+N</kbd> | Create a new window
|<kbd>Alt+F</kbd> | Make the selected window fullscreen. Useful for copying text
|<kbd>Alt+&uarr;/&darr;/&larr;/&rarr;</kbd><br><kbd>Alt+h/j/k/l</kbd> | Select an adjacent window
|<kbd>Alt+Shift+&uarr;/&darr;/&larr;/&rarr;</kbd><br><kbd>Alt+Shift+h/j/k/l</kbd> | Move the selected window
|<kbd>Alt+R</kbd> | Enter resize mode. Resize selected window with arrow keys or <kbd>h/j/k/l</kbd>. Exit using any other key(s)
|<kbd>Alt+/</kbd> | Enter search mode. Type query, navigate between results with arrow keys or <kbd>n/N</kbd>
|<kbd>Scroll</kbd> | Move through scrollback
