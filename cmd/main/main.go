package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/RubikNube/TerminalChess/pkg/gui"
	"github.com/jroimartin/gocui"
)

type Config struct {
	Keybindings map[string]string `json:"keybindings"`
}

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

var (
	board       gui.ChessBoard
	cursor      gui.Cursor
	selectedRow int
	selectedCol int
	selected    bool
)

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("board", 0, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Terminal Chess"
		v.Wrap = false
	}
	if v, err := g.View("board"); err == nil {
		board.RenderToView(v, cursor.Row, cursor.Col, selected, selectedRow, selectedCol)
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
	if !selected && board[cursor.Row][cursor.Col].Type != gui.Empty {
		selected = true
		selectedRow = cursor.Row
		selectedCol = cursor.Col
	}
	return nil
}

func dropPiece(g *gocui.Gui, v *gocui.View) error {
	if selected {
		// Replace with correct move function from ChessBoard
		board.MovePiece(selectedRow, selectedCol, cursor.Row, cursor.Col)
		selected = false
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func reset(g *gocui.Gui, v *gocui.View) error {
	board = gui.NewChessBoard()
	// Reset cursor position
	cursor = gui.Cursor{Row: 0, Col: 0}
	return layout(g)
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

	board = gui.NewChessBoard()
	cursor = gui.Cursor{Row: 0, Col: 0}

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

	g.SetKeybinding("", moveLeftKey, gocui.ModNone, moveCursor(0, -1))
	g.SetKeybinding("", moveRightKey, gocui.ModNone, moveCursor(0, 1))
	g.SetKeybinding("", moveUpKey, gocui.ModNone, moveCursor(-1, 0))
	g.SetKeybinding("", moveDownKey, gocui.ModNone, moveCursor(1, 0))
	g.SetKeybinding("", selectKey, gocui.ModNone, selectPiece)
	g.SetKeybinding("", dropKey, gocui.ModNone, dropPiece)
	g.SetKeybinding("", quitKey, gocui.ModNone, quit)
	g.SetKeybinding("", resetKey, gocui.ModNone, reset)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
