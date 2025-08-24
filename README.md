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

## Loading and Saving

You can save the current game state by pressing `w`. The game will be saved in
the `saves` directory. You can load a game by pressing the `a` key and then selecting
the desired save file. Cancelling the loading dialog is be done by pressing `Ctrl+q`.

By pressing `tab` the loading dialog will autocomplete the folder and file names.
Pressing `tab` multiple times will cycle through the available folders and files.

## Engine

The used default engine is Stockfish. It should be installed on your system in
order to play against it.

The settings for the engine can be changed in the `engine.json` file. If your
Stockfish binary is not in the default path (`/usr/bin/stockfish`), you can
change the path in the `engine.json` file.
