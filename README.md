# TerminalChess

A simple chess UI in the terminal

![Chess board](./images/GUI.gif)

## Navigation

The navigation is Vim based, so you can use the following keys to navigate:

- `h` - move left
- `j` - move down
- `k` - move up
- `l` - move right
- `q` - quit the game
- `r` - reset the game
- `w` - save the game into the `saves` directory
- `p` - pick a piece
- `d` - drop a piece
- `t` - toggle the move history
- `b` - switch the chess board
- `e` - toggle the engine mode
- `y` - move back in the move history
- `x` - move forward in the move history
These defaults can be changed in the config.json file.

## Engine

The used default engine is Stockfish. It should be installed on your system in
order to play against it.
