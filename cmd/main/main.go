package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/RubikNube/TerminalChess/pkg/gui"
	"github.com/RubikNube/TerminalChess/pkg/history"
	"github.com/jroimartin/gocui"
)

type Config struct {
	Keybindings map[string]string `json:"keybindings"`
}

var (
	board        gui.ChessBoard
	cursor       gui.Cursor
	selectedRow  int
	selectedCol  int
	selected     bool
	turn         gui.Color = gui.White // Track whose turn it is
	showHistory  bool      = true      // Track if history view is shown
	enPassantRow int
	enPassantCol int // Track en passant square
)

func loadConfig(path string) (Config, error) {
	var cfg Config
	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&cfg)
	return cfg, err
}

func layout(g *gocui.Gui) error {
	_, maxY := g.Size()
	historyWidth := 20

	// Calculate the exact width needed for the chessboard view
	artWidth := 7
	squareWidth := artWidth*2 + 2
	offset := 13
	boardWidth := 2 + 8*squareWidth - offset // 2 for left label, 8 squares

	// Board view on the left
	if v, err := g.SetView("board", 0, 0, boardWidth-1, maxY-4); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Terminal Chess"
		v.Wrap = false
	}
	if v, err := g.View("board"); err == nil {
		board.RenderToView(v, cursor.Row, cursor.Col, selected, selectedRow, selectedCol)
	}

	// History view on the right (only if showHistory is true)
	if showHistory {
		if v, err := g.SetView("history", boardWidth, 0, boardWidth+historyWidth-1, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "Move History"
			v.Wrap = false
		}
		if v, err := g.View("history"); err == nil {
			v.Clear()
			historyLines := history.GetMoveHistorySAN()
			for _, line := range historyLines {
				fmt.Fprintln(v, line)
			}
		}
	} else {
		// If the view exists but should not be shown, delete it
		if _, err := g.View("history"); err == nil {
			g.DeleteView("history")
		}
	}

	// Info view below the board
	if v, err := g.SetView("info", 0, maxY-3, boardWidth-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = false
	}
	if v, err := g.View("info"); err == nil {
		v.Clear()
		if selected {
			colChar := 'a' + selectedCol
			rowChar := '8' - selectedRow
			fmt.Fprintf(v, "Selected square: %c%c\n", colChar, rowChar)
		} else {
			fmt.Fprint(v, "No square selected\n")
		}
	}
	return nil
}

func moveCursor(dRow, dCol int) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		cursor.Move(dRow, dCol)
		return nil
	}
}

func selectPiece(g *gocui.Gui, v *gocui.View) error {
	if board[cursor.Row][cursor.Col].Type != gui.Empty {
		selected = true
		selectedRow = cursor.Row
		selectedCol = cursor.Col
	}
	return nil
}

func dropPiece(g *gocui.Gui, v *gocui.View) error {
	if selected {
		if board.MovePiece(selectedRow, selectedCol, cursor.Row, cursor.Col, turn) {
			selected = false
			// Switch turn after a successful move
			if turn == gui.White {
				turn = gui.Black
			} else {
				turn = gui.White
			}
		}
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func reset(g *gocui.Gui, v *gocui.View) error {
	board = gui.NewChessBoard()
	// Reset cursor position and turn
	cursor = gui.Cursor{Row: 0, Col: 0}
	turn = gui.White
	history.ClearHistory()
	return layout(g)
}

func toggleHistory(g *gocui.Gui, v *gocui.View) error {
	showHistory = !showHistory
	return nil // layout will be called automatically on next refresh
}

func switchBoard(g *gocui.Gui, v *gocui.View) error {
	gui.ToggleBoardOrientation()
	return nil
}

func moveLeft(g *gocui.Gui, v *gocui.View) error {
	if gui.BoardFlipped {
		return moveCursor(0, 1)(g, v)
	}

	return moveCursor(0, -1)(g, v)
}

func moveRight(g *gocui.Gui, v *gocui.View) error {
	if gui.BoardFlipped {
		return moveCursor(0, -1)(g, v)
	}

	return moveCursor(0, 1)(g, v)
}

func moveUp(g *gocui.Gui, v *gocui.View) error {
	if gui.BoardFlipped {
		return moveCursor(1, 0)(g, v)
	}

	return moveCursor(-1, 0)(g, v)
}

func moveDown(g *gocui.Gui, v *gocui.View) error {
	if gui.BoardFlipped {
		return moveCursor(-1, 0)(g, v)
	}

	return moveCursor(1, 0)(g, v)
}

func main() {
	// Load config
	cfg, err := loadConfig("config.json")
	if err != nil {
		log.Panicln("Failed to load config:", err)
	}
	keybindings := cfg.Keybindings

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	defaultPiecesPath := filepath.Join("assets", "pieces", "default")
	err = gui.LoadAsciiPieces(defaultPiecesPath)
	if err != nil {
		log.Panicln("Failed to load ASCII pieces:", err)
	}
	board = gui.NewChessBoard()
	cursor = gui.Cursor{Row: 6, Col: 4}

	g.SetManagerFunc(layout)

	// Keybindings for movement
	moveLeftKey := []rune(keybindings["moveLeft"])[0]
	moveRightKey := []rune(keybindings["moveRight"])[0]
	moveUpKey := []rune(keybindings["moveUp"])[0]
	moveDownKey := []rune(keybindings["moveDown"])[0]
	selectKey := []rune(keybindings["pick"])[0]
	quitKey := []rune(keybindings["quit"])[0]
	resetKey := []rune(keybindings["reset"])[0]
	dropKey := []rune(keybindings["drop"])[0]
	toggleHistoryKey := []rune(keybindings["toggleHistory"])[0]
	switchBoardKey := []rune(keybindings["switchBoard"])[0]

	g.SetKeybinding("", moveLeftKey, gocui.ModNone, moveLeft)
	g.SetKeybinding("", moveRightKey, gocui.ModNone, moveRight)
	g.SetKeybinding("", moveUpKey, gocui.ModNone, moveUp)
	g.SetKeybinding("", moveDownKey, gocui.ModNone, moveDown)
	g.SetKeybinding("", selectKey, gocui.ModNone, selectPiece)
	g.SetKeybinding("", dropKey, gocui.ModNone, dropPiece)
	g.SetKeybinding("", quitKey, gocui.ModNone, quit)
	g.SetKeybinding("", resetKey, gocui.ModNone, reset)
	g.SetKeybinding("", toggleHistoryKey, gocui.ModNone, toggleHistory)
	g.SetKeybinding("", switchBoardKey, gocui.ModNone, switchBoard)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
