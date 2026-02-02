# TerminalChess

A simple chess UI in the terminal

![Chess board](./images/GUI.gif)

## Navigation

### Basic Navigation

The navigation is Vim based, so you can use the following keys to navigate:

- `h` - move left
- `j` - move down
- `k` - move up
- `l` - move right
- `q` - quit the game
- `r` - reset the game
- `w` - save the game into the `saves` directory
- `W` - copy the current game PGN to the clipboard
- `a` - load a game from the given path
- `p` - pick a piece
- `c` - clears the current selection
- `d` - drop a piece
- `t` - toggle the move history
- `b` - switch the chess board
- `e` - toggle the engine mode
- `y` - move back in the move history
- `x` - move forward in the move history
These defaults can be changed in the config.json file.

### Load dialog navigation

The load dialog can be navigated using the following keys:

- `Ctrl+y` - move up in the auto-complete list
- `Ctrl+x` - move down in the auto-complete list
- `Enter` - select the current file for loading
- `Ctrl+q` - cancel the loading dialog

## Engine

The used default engine is Stockfish. It should be installed on your system in
order to play against it.

The settings for the engine can be changed in the `engine.json` file. If your
Stockfish binary is not in the default path (`/usr/bin/stockfish`), you can
change the path in the `engine.json` file.
