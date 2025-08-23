package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RubikNube/TerminalChess/pkg/engine"
	"github.com/RubikNube/TerminalChess/pkg/gui"
	"github.com/RubikNube/TerminalChess/pkg/history"
	"github.com/corentings/chess"
	"github.com/jroimartin/gocui"
)

type Config struct {
	Keybindings map[string]string `json:"keybindings"`
}

var (
	board             gui.ChessBoard
	cursor            gui.Cursor
	selectedRow       int = -1
	selectedCol       int = -1
	selected          bool
	turn              gui.Color = gui.White // Track whose turn it is
	showHistory       bool      = true      // Track if history view is shown
	enPassantRow      int
	enPassantCol      int    // Track en passant square
	showEngineDialog  bool   = false
	showLoadDialog    bool   = false
	historyIndex      int    = -1 // -1 means current/latest position
	infoMessage       string = "" // Message to show in the info view
	cfg               Config
	defaultLoadPrompt = "Enter path to PGN file:"

	cyclePrefix  string
	cycleIndex   int
	cycleMatches []string
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

	// Render load dialog if needed
	if showLoadDialog {
		loadWidth := 40
		loadHeight := 5
		x := 5
		y := 3
		if v, err := g.SetView("load", x, y, x+loadWidth, y+loadHeight); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "Load Game"
			v.Wrap = true
			v.Editable = true
			v.Clear()
			fmt.Fprintln(v, "Enter path to PGN file:")
			g.SetCurrentView("load")
			g.Cursor = true
			v.Editor = &loadEditor{defaultPrompt: defaultLoadPrompt}
		}
	} else {
		// Remove load dialog if it exists
		if _, err := g.View("load"); err == nil {
			g.DeleteView("load")
		}
	}

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
		// Show board at selected history index if navigating
		if historyIndex >= 0 {
			hist := history.GetHistory()
			game := chess.NewGame()
			for i := 0; i <= historyIndex && i < len(hist); i++ {
				move, err := chess.UCINotation{}.Decode(game.Position(), hist[i])
				if err == nil {
					game.Move(move)
				}
			}
			fen := game.Position().Board().String()
			tmpBoard := gui.NewChessBoardFromFEN(fen)
			tmpBoard.RenderToView(v, cursor.Row, cursor.Col, selected, selectedRow, selectedCol)
		} else {
			board.RenderToView(v, cursor.Row, cursor.Col, selected, selectedRow, selectedCol)
		}
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
			for i, line := range historyLines {
				if historyIndex >= 0 && i == historyIndex/2 {
					fmt.Fprintf(v, "> %s\n", line)
				} else {
					fmt.Fprintln(v, line)
				}
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

	return nil
}

// Helper to show a message in the InfoView
func showInfoMessage(g *gocui.Gui, msg string) {
	if v, err := g.View("info"); err == nil {
		v.Clear()
		fmt.Fprintln(v, msg)
	}
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
	if selected && selectedRow >= 0 && selectedCol >= 0 {
		if board.MovePiece(selectedRow, selectedCol, cursor.Row, cursor.Col, turn) {
			selected = false
			if turn == gui.White {
				turn = gui.Black
			} else {
				turn = gui.White
			}
			// If automove is enabled and it's now the engine's turn, trigger engine move
			if engine.LoadedEngineConfig.Automove && ((engine.LoadedEngineConfig.EngineColor == "white" && turn == gui.White) || (engine.LoadedEngineConfig.EngineColor == "black" && turn == gui.Black)) {
				engineMove(g, v)
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

func openLoadDialog(g *gocui.Gui, v *gocui.View) error {
	showLoadDialog = true
	enableLoadDialogKeybindings(g)
	return layout(g)
}

func clearSelection(g *gocui.Gui, v *gocui.View) error {
	selected = false
	selectedRow = -1
	selectedCol = -1
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

func engineMove(g *gocui.Gui, v *gocui.View) error {
	fen := board.ToFEN(turn)
	bestMove, err := engine.GetBestMove(fen, 10)
	if err != nil || bestMove == "" {
		log.Println("Error: Could not get best move from Stockfish.")
		return nil
	}
	if len(bestMove) < 4 {
		log.Println("Error: Invalid best move format.")
		return nil
	}
	fromCol := int(bestMove[0] - 'a')
	fromRow := 8 - int(bestMove[1]-'0')
	toCol := int(bestMove[2] - 'a')
	toRow := 8 - int(bestMove[3]-'0')
	if board.MovePiece(fromRow, fromCol, toRow, toCol, turn) {
		// Switch turn after a successful move
		if turn == gui.White {
			turn = gui.Black
		} else {
			turn = gui.White
		}
	}
	return nil
}

func historyPrev(g *gocui.Gui, v *gocui.View) error {
	hist := history.GetHistory()
	if len(hist) == 0 {
		return nil
	}
	if historyIndex < 0 {
		historyIndex = len(hist) - 1
	} else if historyIndex > 0 {
		historyIndex--
	}
	return nil
}

func historyNext(g *gocui.Gui, v *gocui.View) error {
	hist := history.GetHistory()
	if len(hist) == 0 {
		return nil
	}
	if historyIndex >= 0 {
		historyIndex++
		if historyIndex >= len(hist) {
			historyIndex = -1
		}
	}
	return nil
}

// Save the current game as a PGN file in the "saves" directory and show notification in InfoView
func saveGameAsPGN(g *gocui.Gui, v *gocui.View) error {
	saveDir := "saves"
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		showInfoMessage(g, fmt.Sprintf("Error creating saves directory: %v", err))
		return nil
	}
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("chess_%s.pgn", timestamp)
	filepath := filepath.Join(saveDir, filename)

	hist := history.GetHistory()
	game := chess.NewGame()
	for _, moveStr := range hist {
		move, err := chess.UCINotation{}.Decode(game.Position(), moveStr)
		if err == nil {
			game.Move(move)
		}
	}

	playerName := os.Getenv("USER")
	if playerName == "" {
		playerName = "Player"
	}
	date := time.Now().Format("2006.01.02")

	elo := 0
	if eloOpt, ok := engine.LoadedEngineConfig.Options["UCI_Elo"]; ok {
		switch v := eloOpt.(type) {
		case float64:
			elo = int(v)
		case int:
			elo = v
		}
	}

	f, err := os.Create(filepath)
	if err != nil {
		showInfoMessage(g, fmt.Sprintf("Error creating PGN file: %v", err))
		return nil
	}
	defer f.Close()

	fmt.Fprintf(f, "[Event \"Casual Game\"]\n")
	fmt.Fprintf(f, "[Date \"%s\"]\n", date)
	// determine if the engine is playing white or black
	if engine.LoadedEngineConfig.EngineColor == "white" {
		fmt.Fprintf(f, "[White \"%s (Elo: %d)\"]\n", engine.LoadedEngineConfig.Name, elo)
		fmt.Fprintf(f, "[Black \"%s\"]\n", playerName)
	} else {
		fmt.Fprintf(f, "[White \"%s\"]\n", playerName)
		fmt.Fprintf(f, "[Black \"%s (Elo: %d)\"]\n", engine.LoadedEngineConfig.Name, elo)
	}
	fmt.Fprintf(f, "\n")

	line := game.String()
	if line != "" {
		fmt.Fprintln(f, line)
	}

	notification := fmt.Sprintf("Game saved to saves/%s", filename)
	showInfoMessage(g, notification)
	return nil
}

func handleLoadGame(g *gocui.Gui, v *gocui.View) error {
	buf := v.Buffer()
	lines := strings.Split(buf, "\n")
	if len(lines) < 2 {
		showInfoMessage(g, "Please enter a valid path.")
		return nil
	}
	path := strings.TrimSpace(lines[1])
	f, err := os.Open(path)
	if err != nil {
		showInfoMessage(g, fmt.Sprintf("Failed to open file: %v", err))
		return nil
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		showInfoMessage(g, fmt.Sprintf("Failed to read file: %v", err))
		return nil
	}
	gameFunc, err := chess.PGN(strings.NewReader(string(data)))
	if err != nil {
		showInfoMessage(g, "Invalid PGN file.")
		return nil
	}
	game := chess.NewGame()
	gameFunc(game)
	parsedGame := game
	if parsedGame == nil {
		showInfoMessage(g, "Invalid PGN file.")
		return nil
	}
	history.ClearHistory()
	for _, move := range parsedGame.Moves() {
		history.AddMove(chess.UCINotation{}.Encode(parsedGame.Position(), move))
	}
	board = gui.NewChessBoardFromFEN(parsedGame.FEN())
	showLoadDialog = false
	g.DeleteView("load")
	g.SetCurrentView("board")
	showInfoMessage(g, fmt.Sprintf("Loaded game from %s", path))
	return layout(g)
}

// Autocomplete file path in load dialog when Tab is pressed
func autocompleteFilePath(g *gocui.Gui, v *gocui.View) error {
	buf := v.Buffer()
	input := strings.TrimSpace(buf)
	dir, filePrefix := filepath.Split(input)
	if dir == "" {
		dir = "."
	}

	// Only recalculate matches if prefix changed or no matches cached
	if filePrefix != cyclePrefix || cycleMatches == nil {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}
		var matches []string
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasPrefix(name, filePrefix) {
				matches = append(matches, name)
			}
		}
		cyclePrefix = filePrefix
		cycleMatches = matches
		cycleIndex = 0
	}

	if len(cycleMatches) == 1 {
		// Single match, autocomplete directly
		input = filepath.Join(dir, cycleMatches[0])
		cycleMatches = nil
		cycleIndex = 0
	} else if len(cycleMatches) > 1 {
		// Cycle through cached matches only, do not recalculate
		input = filepath.Join(dir, cycleMatches[cycleIndex])
		showInfoMessage(g, "Matches: "+strings.Join(cycleMatches, " "))
		cycleIndex = (cycleIndex + 1) % len(cycleMatches)
	} else {
		cycleMatches = nil
		cycleIndex = 0
	}

	v.Clear()
	fmt.Fprint(v, input)
	return nil
}

func enableGlobalKeybindings(g *gocui.Gui, keybindings map[string]string) {
	g.DeleteKeybindings("")
	g.DeleteKeybindings("load")
	moveLeftKey := []rune(keybindings["moveLeft"])[0]
	moveRightKey := []rune(keybindings["moveRight"])[0]
	moveUpKey := []rune(keybindings["moveUp"])[0]
	moveDownKey := []rune(keybindings["moveDown"])[0]
	selectKey := []rune(keybindings["pick"])[0]
	clearSelectionKey := []rune(keybindings["clearSelection"])[0]
	quitKey := []rune(keybindings["quit"])[0]
	resetKey := []rune(keybindings["reset"])[0]
	dropKey := []rune(keybindings["drop"])[0]
	toggleHistoryKey := []rune(keybindings["toggleHistory"])[0]
	switchBoardKey := []rune(keybindings["switchBoard"])[0]
	engineMoveKey := []rune(keybindings["engineMove"])[0]
	forwardHistoryKey := []rune(keybindings["historyForward"])[0]
	backwardHistoryKey := []rune(keybindings["historyBackward"])[0]
	saveGameKey := []rune(keybindings["saveGame"])[0]
	loadGameKey := []rune(keybindings["loadGame"])[0]

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
	g.SetKeybinding("", engineMoveKey, gocui.ModNone, engineMove)
	g.SetKeybinding("", backwardHistoryKey, gocui.ModNone, historyPrev)
	g.SetKeybinding("", forwardHistoryKey, gocui.ModNone, historyNext)
	g.SetKeybinding("", saveGameKey, gocui.ModNone, saveGameAsPGN)
	g.SetKeybinding("", clearSelectionKey, gocui.ModNone, clearSelection)
	g.SetKeybinding("", loadGameKey, gocui.ModNone, openLoadDialog)
}

func enableLoadDialogKeybindings(g *gocui.Gui) {
	g.DeleteKeybindings("")
	g.DeleteKeybindings("load")
	g.SetKeybinding("load", gocui.KeyEnter, gocui.ModNone, handleLoadGame)
	g.SetKeybinding("load", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		showLoadDialog = false
		enableGlobalKeybindings(g, cfg.Keybindings)
		g.DeleteView("load")
		g.SetCurrentView("board")
		return layout(g)
	})
	g.SetKeybinding("load", gocui.KeyCtrlQ, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		showLoadDialog = false
		enableGlobalKeybindings(g, cfg.Keybindings)
		g.DeleteView("load")
		g.SetCurrentView("board")
		return layout(g)
	})
	g.SetKeybinding("load", gocui.KeyTab, gocui.ModNone, autocompleteFilePath)
	// Remove default prompt when user types any character
	g.SetKeybinding("load", 0, gocui.ModNone, clearLoadPromptOnRune)
}

func clearLoadPromptOnRune(g *gocui.Gui, v *gocui.View) error {
	buf := v.Buffer()
	lines := strings.Split(buf, "\n")
	// If the prompt is present, clear it and leave only an empty input line
	if len(lines) > 0 && strings.Contains(lines[0], defaultLoadPrompt) {
		v.Clear()
		fmt.Fprintln(v, "")
	}
	return nil
}

type loadEditor struct {
	defaultPrompt string
	cleared       bool
}

func (e *loadEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if !e.cleared {
		buf := v.Buffer()
		lines := strings.Split(buf, "\n")
		if len(lines) > 0 && strings.Contains(lines[0], e.defaultPrompt) {
			v.Clear()
			e.cleared = true
		}
	}
	gocui.DefaultEditor.Edit(v, key, ch, mod)
}

func clearLoadPromptOnInput(g *gocui.Gui, v *gocui.View) error {
	buf := v.Buffer()
	lines := strings.Split(buf, "\n")
	if len(lines) > 0 && strings.Contains(lines[0], defaultLoadPrompt) {
		v.Clear()
		fmt.Fprintln(v, "")
		if len(lines) > 1 {
			fmt.Fprintln(v, lines[1])
		}
	}
	return nil
}

func main() {
	// Load config
	var err error
	cfg, err = loadConfig("config.json")
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

	engine.Initialize("engine.json")

	enableGlobalKeybindings(g, keybindings)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
