## Rendering
Each time a `Split` is modified, its elements' placement on the screen is recalculated.

Rendering placement is updated recursively.

Terminals should always be able to move their contents to another portion of the screen. This means that they must be able to redraw their buffer, suggesting that they should maintain a buffer themselves instead of relying on the host terminal's buffer.

The buffer will not be a 2d array of runes because we want to preserve color.
Rather, the buffer should perhaps just be a string that would render properly at the top-left corner of the screen. This buffer can then be properly translated to the appropriate portion of the screen by rewriting CSI.

## Rendering a window

### Incrememtal Updates

### Force Redraws
