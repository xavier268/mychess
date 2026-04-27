package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xavier268/mychess"
	"github.com/xavier268/mychess/game"
	"github.com/xavier268/mychess/position"
)

const analysisDepth = 50

//go:embed static/index.html
var indexHTML []byte

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ── WebSocket Hub ─────────────────────────────────────────────────────────────

type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

type hub struct {
	clients    map[*wsClient]bool
	broadcast  chan []byte
	register   chan *wsClient
	unregister chan *wsClient
}

func newHub() *hub {
	return &hub{
		clients:    make(map[*wsClient]bool),
		broadcast:  make(chan []byte, 16),
		register:   make(chan *wsClient),
		unregister: make(chan *wsClient),
	}
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					close(c.send)
					delete(h.clients, c)
				}
			}
		}
	}
}

// ── Server ────────────────────────────────────────────────────────────────────

type Server struct {
	g          *game.Game
	ctx        context.Context
	cancel     context.CancelFunc
	h          *hub
	httpServer *http.Server

	moveMu sync.Mutex // serialises move, reset, and shutdown operations

	mu          sync.Mutex // protects display state and s.g pointer below
	displayPos  position.Position
	historyStr  string
	whiteTime   time.Duration
	blackTime   time.Duration
	turnStart   time.Time
	gameOver    bool
	message     string
	problemMode bool
}

func newServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())
	g := game.NewGame()
	s := &Server{
		g:          g,
		ctx:        ctx,
		cancel:     cancel,
		h:          newHub(),
		displayPos: g.Position,
		turnStart:  time.Now(),
	}
	go s.h.run()
	g.AnalysisAsync(ctx, analysisDepth)
	go s.tickerLoop()
	return s
}

func (s *Server) run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/ws", s.handleWS)
	s.httpServer = &http.Server{Addr: addr, Handler: mux}
	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

// ── HTTP Handlers ─────────────────────────────────────────────────────────────

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &wsClient{conn: conn, send: make(chan []byte, 8)}
	s.h.register <- c
	c.send <- s.buildState() // initial state in buffer
	go s.writePump(c)
	s.readPump(c) // blocks until disconnect
}

func (s *Server) writePump(c *wsClient) {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (s *Server) readPump(c *wsClient) {
	defer func() { s.h.unregister <- c }()
	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		var msg struct {
			Type         string `json:"type"`
			From         string `json:"from"`
			To           string `json:"to"`
			Promotion    string `json:"promotion"`
			Fen          string `json:"fen"`
			Turn         string `json:"turn"`
			CastleWhiteK bool   `json:"castleWhiteK"`
			CastleWhiteQ bool   `json:"castleWhiteQ"`
			CastleBlackK bool   `json:"castleBlackK"`
			CastleBlackQ bool   `json:"castleBlackQ"`
		}
		if json.Unmarshal(raw, &msg) != nil {
			continue
		}
		switch msg.Type {
		case "move":
			s.handleMove(msg.From, msg.To, msg.Promotion)
		case "auto":
			s.handleAuto()
		case "reset":
			s.handleReset()
		case "quit":
			s.handleShutdown()
		case "problem_enter":
			s.handleProblemEnter()
		case "problem_exit":
			s.handleProblemExit(msg.Fen, msg.Turn, msg.CastleWhiteK, msg.CastleWhiteQ, msg.CastleBlackK, msg.CastleBlackQ)
		}
	}
}

// ── Move Handling ─────────────────────────────────────────────────────────────

func (s *Server) handleMove(from, to, promo string) {
	s.moveMu.Lock()
	defer s.moveMu.Unlock()

	s.mu.Lock()
	if s.gameOver || s.problemMode {
		s.mu.Unlock()
		return
	}
	pos := s.displayPos
	s.mu.Unlock()

	mv, err := parseMove(from+to+promo, pos)
	if err != nil {
		s.mu.Lock()
		s.message = err.Error()
		s.mu.Unlock()
		s.broadcastState()
		return
	}

	s.mu.Lock()
	if pos.Turn() == position.WHITE {
		s.whiteTime += time.Since(s.turnStart)
	} else {
		s.blackTime += time.Since(s.turnStart)
	}
	s.turnStart = time.Now()
	s.mu.Unlock()

	s.g.Play(mv)

	s.mu.Lock()
	s.displayPos = s.g.Position
	s.historyStr = buildHistoryStr(s.g.History)
	gameOver, msg := s.checkGameOver()
	s.gameOver = gameOver
	if gameOver {
		s.message = msg
	} else {
		s.message = "joué : " + from + to + promo
	}
	s.mu.Unlock()

	if !gameOver {
		s.g.Z.ResetStats()
		s.g.AnalysisAsync(s.ctx, analysisDepth)
	}
	s.broadcastState()
}

func (s *Server) handleAuto() {
	s.moveMu.Lock()
	defer s.moveMu.Unlock()

	s.mu.Lock()
	if s.gameOver || s.problemMode {
		s.mu.Unlock()
		return
	}
	turn := s.displayPos.Turn()
	s.mu.Unlock()

	if err := s.g.AutoPlay(); err != nil {
		s.mu.Lock()
		s.message = "autoplay : " + err.Error()
		s.mu.Unlock()
		s.broadcastState()
		return
	}

	s.mu.Lock()
	if turn == position.WHITE {
		s.whiteTime += time.Since(s.turnStart)
	} else {
		s.blackTime += time.Since(s.turnStart)
	}
	s.turnStart = time.Now()
	s.displayPos = s.g.Position
	s.historyStr = buildHistoryStr(s.g.History)
	gameOver, msg := s.checkGameOver()
	s.gameOver = gameOver
	if gameOver {
		s.message = msg
	} else {
		s.message = "coup automatique joué"
	}
	s.mu.Unlock()

	if !gameOver {
		s.g.Z.ResetStats()
		s.g.AnalysisAsync(s.ctx, analysisDepth)
	}
	s.broadcastState()
}

// ── Reset & Shutdown ──────────────────────────────────────────────────────────

func (s *Server) handleReset() {
	s.moveMu.Lock()
	defer s.moveMu.Unlock()

	// Signal current analysis to stop, then wait for the goroutine to release
	// g.mu (AnalysisAsync acquires g.mu in the calling goroutine before returning).
	stopCtx, stopCancel := context.WithCancel(context.Background())
	stopCancel() // pre-cancelled: any new analysis exits immediately
	s.g.AnalysisAsync(stopCtx, 1)

	// Show loading feedback while the cache is being read from disk.
	s.mu.Lock()
	s.message = "Chargement de la table…"
	s.mu.Unlock()
	s.broadcastState()

	newG := game.NewGame()

	s.mu.Lock()
	s.g = newG
	s.displayPos = newG.Position
	s.historyStr = ""
	s.whiteTime = 0
	s.blackTime = 0
	s.turnStart = time.Now()
	s.gameOver = false
	s.message = "Nouvelle partie"
	s.mu.Unlock()

	newG.AnalysisAsync(s.ctx, analysisDepth)
	s.broadcastState()
}

func (s *Server) handleShutdown() {
	s.mu.Lock()
	s.message = "Serveur en cours d'arrêt…"
	s.mu.Unlock()
	s.broadcastState()

	go func() {
		time.Sleep(300 * time.Millisecond) // let the broadcast reach clients
		s.cancel()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx) //nolint:errcheck
	}()
}

// ── Problem Mode ─────────────────────────────────────────────────────────────

func (s *Server) handleProblemEnter() {
	s.moveMu.Lock()
	defer s.moveMu.Unlock()

	// Stop any running analysis immediately.
	stopCtx, stopCancel := context.WithCancel(context.Background())
	stopCancel()
	s.g.AnalysisAsync(stopCtx, 1)

	s.mu.Lock()
	s.problemMode = true
	s.message = "Mode problème — édition de la position"
	s.mu.Unlock()
	s.broadcastState()
}

func (s *Server) handleProblemExit(fen, turn string, castleWK, castleWQ, castleBK, castleBQ bool) {
	s.moveMu.Lock()
	defer s.moveMu.Unlock()

	pos, err := position.ParseFEN(fen)
	if err != nil {
		s.mu.Lock()
		s.message = "Position invalide : " + err.Error()
		s.mu.Unlock()
		s.broadcastState()
		return
	}

	if msg := pos.Validate(); msg != "" {
		s.mu.Lock()
		s.message = "Position invalide : " + msg
		s.mu.Unlock()
		s.broadcastState()
		return
	}

	// Both kings must be present on the board.
	if pos.PieceAt(pos.KingPosition(position.WHITE)) != position.KING {
		s.mu.Lock()
		s.message = "Position invalide : roi blanc absent"
		s.mu.Unlock()
		s.broadcastState()
		return
	}
	if pos.PieceAt(pos.KingPosition(position.BLACK)) != -position.KING {
		s.mu.Lock()
		s.message = "Position invalide : roi noir absent"
		s.mu.Unlock()
		s.broadcastState()
		return
	}

	if turn == "b" {
		pos.SetTurn(position.BLACK)
	} else {
		pos.SetTurn(position.WHITE)
	}

	var castleW, castleB uint8
	if castleWK {
		castleW |= position.CanCastleKingSide
	}
	if castleWQ {
		castleW |= position.CanCastleQueenSide
	}
	if castleBK {
		castleB |= position.CanCastleKingSide
	}
	if castleBQ {
		castleB |= position.CanCastleQueenSide
	}
	pos.SetCastle(position.WHITE, castleW)
	pos.SetCastle(position.BLACK, castleB)
	pos.ResetEnPassantFlag()
	pos.Hash = position.DefaultZT.HashPosition(pos)

	s.g.SetPosition(pos)

	s.mu.Lock()
	s.problemMode = false
	s.displayPos = pos
	s.historyStr = ""
	s.whiteTime = 0
	s.blackTime = 0
	s.turnStart = time.Now()
	s.gameOver = false
	s.message = "Position éditée — analyse en cours…"
	s.mu.Unlock()

	s.g.AnalysisAsync(s.ctx, analysisDepth)
	s.broadcastState()
}

// ── Game Over Detection ───────────────────────────────────────────────────────

// called while s.mu is held; s.g.History is stable because moveMu is held by the caller.
func (s *Server) checkGameOver() (bool, string) {
	if len(s.displayPos.GetMoveList()) == 0 {
		if s.displayPos.IsCheck() {
			winner := "Noirs"
			if s.displayPos.Turn() != position.WHITE {
				winner = "Blancs"
			}
			return true, fmt.Sprintf("ÉCHEC ET MAT — Les %s gagnent !", winner)
		}
		return true, "PAT — Partie nulle !"
	}
	pm := make(map[uint64]int, len(s.g.History)+1)
	for _, move := range s.g.History {
		pm[move.PrevHash]++
		if pm[move.PrevHash] >= 3 {
			return true, "PAT par 3 répétitions — Partie nulle !"
		}
	}
	return false, ""
}

// ── State Building & Broadcast ────────────────────────────────────────────────

type StateMsg struct {
	Type        string  `json:"type"`
	Fen         string  `json:"fen"` // piece-placement FEN, used by chessboard.js
	Turn        string  `json:"turn"`
	Score       int     `json:"score"`
	History     string  `json:"history"`
	Depth       int     `json:"depth"`
	Best        string  `json:"best"`
	TableStats  string  `json:"tableStats"`
	WhiteTime   float64 `json:"whiteTime"`
	BlackTime   float64 `json:"blackTime"`
	GameOver    bool    `json:"gameOver"`
	Message     string  `json:"message"`
	Version     string  `json:"version"`
	ProblemMode bool    `json:"problemMode"`
	CastleWK    bool    `json:"castleWK"`
	CastleWQ    bool    `json:"castleWQ"`
	CastleBK    bool    `json:"castleBK"`
	CastleBQ    bool    `json:"castleBQ"`
}

func (s *Server) buildState() []byte {
	s.mu.Lock()
	g := s.g // capture pointer so a concurrent reset can't swap it mid-build
	pos := s.displayPos
	wt := s.whiteTime
	bt := s.blackTime
	ts := s.turnStart
	gameOver := s.gameOver
	message := s.message
	historyStr := s.historyStr
	problemMode := s.problemMode
	s.mu.Unlock()

	if !gameOver {
		if pos.Turn() == position.WHITE {
			wt += time.Since(ts)
		} else {
			bt += time.Since(ts)
		}
	}

	// Read analysis snapshot from captured g (display-only tolerance, same as bubbletea client).
	entry := g.LastRootEntry
	best := "–"
	if entry.Best.From != entry.Best.To || entry.Best.Promotion != position.EMPTY {
		best = moveStr(entry.Best)
	}

	turn := "Blancs"
	if pos.Turn() == position.BLACK {
		turn = "Noirs"
	}

	data, _ := json.Marshal(StateMsg{
		Type:        "state",
		Fen:         positionToFEN(pos),
		Turn:        turn,
		Score:       int(entry.Score),
		History:     historyStr,
		Depth:       int(entry.Depth),
		Best:        best,
		TableStats:  g.Z.Stats(),
		WhiteTime:   wt.Seconds(),
		BlackTime:   bt.Seconds(),
		GameOver:    gameOver,
		Message:     message,
		Version:     mychess.VERSION,
		ProblemMode: problemMode,
		CastleWK:    pos.CastleBits(position.WHITE)&position.CanCastleKingSide != 0,
		CastleWQ:    pos.CastleBits(position.WHITE)&position.CanCastleQueenSide != 0,
		CastleBK:    pos.CastleBits(position.BLACK)&position.CanCastleKingSide != 0,
		CastleBQ:    pos.CastleBits(position.BLACK)&position.CanCastleQueenSide != 0,
	})
	return data
}

func (s *Server) broadcastState() {
	s.h.broadcast <- s.buildState()
}

func (s *Server) tickerLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.broadcastState()
		case <-s.ctx.Done():
			return
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// positionToFEN generates the piece-placement section of a FEN string.
// PieceRepresentation already uses uppercase for white and lowercase for black,
// which matches the FEN convention exactly.
func positionToFEN(pos position.Position) string {
	var sb strings.Builder
	for rank := 7; rank >= 0; rank-- {
		empty := 0
		for file := 0; file < 8; file++ {
			p := pos.PieceAt(position.Sq(rank, file))
			if p == position.EMPTY {
				empty++
			} else {
				if empty > 0 {
					sb.WriteByte(byte('0' + empty))
					empty = 0
				}
				sb.WriteRune(position.PieceRepresentation[p])
			}
		}
		if empty > 0 {
			sb.WriteByte(byte('0' + empty))
		}
		if rank > 0 {
			sb.WriteByte('/')
		}
	}
	return sb.String()
}

func moveStr(mv position.Move) string {
	s := mv.From.String() + mv.To.String()
	switch mv.Promotion {
	case position.QUEEN:
		s += "q"
	case position.ROOK:
		s += "r"
	case position.BISHOP:
		s += "b"
	case position.KNIGHT:
		s += "n"
	}
	return s
}

func parseMove(s string, pos position.Position) (position.Move, error) {
	if len(s) < 4 || len(s) > 5 {
		return position.Move{}, fmt.Errorf("coup invalide %q", s)
	}
	from, err := parseSq(s[0:2])
	if err != nil {
		return position.Move{}, fmt.Errorf("case de départ invalide dans %q", s)
	}
	to, err := parseSq(s[2:4])
	if err != nil {
		return position.Move{}, fmt.Errorf("case d'arrivée invalide dans %q", s)
	}
	var promo position.Piece
	if len(s) == 5 {
		switch s[4] {
		case 'q':
			promo = position.QUEEN
		case 'r':
			promo = position.ROOK
		case 'b':
			promo = position.BISHOP
		case 'n':
			promo = position.KNIGHT
		default:
			return position.Move{}, fmt.Errorf("promotion invalide : '%c'", s[4])
		}
	}
	for _, mv := range pos.GetMoveList() {
		if mv.From == from && mv.To == to {
			if promo == position.EMPTY || mv.Promotion == promo {
				return mv, nil
			}
		}
	}
	return position.Move{}, fmt.Errorf("coup illégal : %s", s)
}

func parseSq(s string) (position.Square, error) {
	if len(s) != 2 || s[0] < 'a' || s[0] > 'h' || s[1] < '1' || s[1] > '8' {
		return 0, fmt.Errorf("case invalide : %q", s)
	}
	return position.Sq(int(s[1]-'1'), int(s[0]-'a')), nil
}

func buildHistoryStr(history []position.Move) string {
	if len(history) == 0 {
		return "–"
	}
	var sb strings.Builder
	for i := 0; i < len(history); i += 2 {
		fmt.Fprintf(&sb, "%2d. %s", i/2+1, moveStr(history[i]))
		if i+1 < len(history) {
			fmt.Fprintf(&sb, " %s", moveStr(history[i+1]))
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}
