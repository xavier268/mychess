package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"mychess/game"
	"mychess/position"
)

// analysisDepth est une borne haute intentionnellement grande : en pratique
// l'analyse est toujours interrompue par un coup du joueur avant d'atteindre
// cette profondeur, ce qui lui permet de tourner en continu.
const analysisDepth = 50

// Pièces unicode : blanches (positives) et noires (négatives).
var unicodePiece = map[position.Piece]string{
	position.KING:    "♔",
	position.QUEEN:   "♕",
	position.ROOK:    "♖",
	position.BISHOP:  "♗",
	position.KNIGHT:  "♘",
	position.PAWN:    "♙",
	-position.KING:   "♚",
	-position.QUEEN:  "♛",
	-position.ROOK:   "♜",
	-position.BISHOP: "♝",
	-position.KNIGHT: "♞",
	-position.PAWN:   "♟",
	position.EMPTY:   " ",
}

var (
	lightSqBg = lipgloss.Color("#F0D9B5")
	darkSqBg  = lipgloss.Color("#B58863")
	whiteFg   = lipgloss.Color("#FFFFFF")
	blackFg   = lipgloss.Color("#1A1A1A")

	infoStyle = lipgloss.NewStyle().Padding(0, 1)
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#CC3333")).Bold(true)
	okStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#228833"))
	boldStyle = lipgloss.NewStyle().Bold(true)
)

// ── Messages BubbleTea ────────────────────────────────────────────────────────

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// ── Modèle ────────────────────────────────────────────────────────────────────

type model struct {
	g       *game.Game
	ctx     context.Context
	cancel  context.CancelFunc
	input   string
	message string
	msgOK   bool
	mem     uint64
	zSize   int
	running bool

	// displayPos est une copie stable de la position, mise à jour uniquement
	// lorsqu'un coup est joué ou annulé — jamais lors du tick.
	// Cela évite que l'échiquier clignote pendant l'analyse (AlphaBeta
	// modifie g.Position en continu via DoMove/UndoMove).
	displayPos position.Position
}

func initialModel() model {
	ctx, cancel := context.WithCancel(context.Background())
	g := game.NewGame(ctx)
	m := model{g: g, ctx: ctx, cancel: cancel, displayPos: g.Position}
	g.AnalysisAsync(ctx, analysisDepth)
	return m
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.mem = availableMemoryBytes()
		m.zSize = m.g.ZTableSize()
		m.running = m.g.IsAnalysisRunning()
		return m, tick()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.cancel()
			return m, tea.Quit
		case tea.KeyEnter:
			return m.handleInput()
		case tea.KeyBackspace, tea.KeyDelete:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if msg.Type == tea.KeyRunes {
				m.input += string(msg.Runes)
			}
		}
	}
	return m, nil
}

func (m model) handleInput() (tea.Model, tea.Cmd) {
	raw := strings.TrimSpace(strings.ToLower(m.input))
	m.input = ""

	switch raw {
	case "q", "quit":
		m.cancel()
		return m, tea.Quit

	case "a", "auto":
		if err := m.g.AutoPlay(); err != nil {
			m.message = err.Error()
			m.msgOK = false
		} else {
			m.displayPos = m.g.Position
			m.message = "AutoPlay : coup joué"
			m.msgOK = true
			m.g.AnalysisAsync(m.ctx, analysisDepth)
		}

	case "u", "undo":
		if err := m.g.RetractPlay(); err != nil {
			m.message = err.Error()
			m.msgOK = false
		} else {
			m.displayPos = m.g.Position
			m.message = "Coup annulé"
			m.msgOK = true
			m.g.AnalysisAsync(m.ctx, analysisDepth)
		}

	default:
		mv, err := parseMove(raw, m.displayPos)
		if err != nil {
			m.message = err.Error()
			m.msgOK = false
		} else {
			m.g.Play(mv)
			m.displayPos = m.g.Position
			m.message = fmt.Sprintf("Joué : %s→%s", mv.From, mv.To)
			m.msgOK = true
			m.g.AnalysisAsync(m.ctx, analysisDepth)
		}
	}
	return m, nil
}

// ── Vue ───────────────────────────────────────────────────────────────────────

func (m model) View() string {
	var sb strings.Builder
	sb.WriteString(renderBoard(m.displayPos))
	sb.WriteString("\n")
	sb.WriteString(renderInfo(m))
	sb.WriteString("\n\n")
	sb.WriteString(renderInput(m))
	return sb.String()
}

func renderBoard(pos position.Position) string {
	var sb strings.Builder

	var header strings.Builder
	header.WriteString("  ")
	for f := range 8 {
		fmt.Fprintf(&header, "  %c ", 'a'+f)
	}
	sb.WriteString(boldStyle.Render(header.String()) + "\n")

	for r := 7; r >= 0; r-- {
		sb.WriteString(boldStyle.Render(fmt.Sprintf("%d ", r+1)))
		for f := 0; f < 8; f++ {
			sq := position.Sq(r, f)
			piece := pos.PieceAt(sq)
			light := (r+f)%2 == 0

			sym, ok := unicodePiece[piece]
			if !ok {
				sym = " "
			}
			cell := fmt.Sprintf(" %s  ", sym) // 4 chars wide for alignment

			bg := darkSqBg
			if light {
				bg = lightSqBg
			}
			fg := blackFg
			if piece > 0 {
				fg = whiteFg
			}

			style := lipgloss.NewStyle().Background(bg).Foreground(fg)
			sb.WriteString(style.Render(cell))
		}
		sb.WriteString(boldStyle.Render(fmt.Sprintf(" %d", r+1)) + "\n")
	}

	var footer strings.Builder
	footer.WriteString("  ")
	for f := range 8 {
		fmt.Fprintf(&footer, "  %c ", 'a'+f)
	}
	sb.WriteString(boldStyle.Render(footer.String()) + "\n")

	return sb.String()
}

func renderInfo(m model) string {
	turn := "Blancs"
	if m.displayPos.Turn() == position.BLACK {
		turn = "Noirs"
	}

	scoreStr := "—"
	if entry, ok := m.g.LastRootEntry(); ok {
		switch {
		case entry.Score >= 29000:
			scoreStr = okStyle.Render("Mat annoncé !")
		case entry.Score <= -29000:
			scoreStr = errStyle.Render("Position perdue")
		default:
			pawn := float64(entry.Score) / 10.0
			sign := "+"
			if pawn < 0 {
				sign = ""
			}
			scoreStr = fmt.Sprintf("%s%.1f p.", sign, pawn)
		}
	}

	analysisStr := "⏸  arrêtée"
	if m.running {
		analysisStr = okStyle.Render("⚙  en cours")
	}

	memMB := float64(m.mem) / 1024 / 1024
	memStr := fmt.Sprintf("%.0f Mo", memMB)

	line1 := infoStyle.Render(fmt.Sprintf(
		"Tour : %s   Score : %s   Analyse : %s",
		boldStyle.Render(turn), scoreStr, analysisStr,
	))
	line2 := infoStyle.Render(fmt.Sprintf(
		"Table Z : %d entrées   RAM dispo : %s",
		m.zSize, memStr,
	))

	var sb strings.Builder
	sb.WriteString(line1 + "\n")
	sb.WriteString(line2)

	if m.message != "" {
		sb.WriteString("\n")
		if m.msgOK {
			sb.WriteString(okStyle.Render("✓ " + m.message))
		} else {
			sb.WriteString(errStyle.Render("✗ " + m.message))
		}
	}

	return sb.String()
}

func renderInput(m model) string {
	prompt := boldStyle.Render("Coup")
	hint := lipgloss.NewStyle().Faint(true).Render("(ex: e2e4, e7e8q, auto, undo, q)")
	return fmt.Sprintf("%s %s > %s█", prompt, hint, m.input)
}

// ── Parsing de coup ───────────────────────────────────────────────────────────

// parseMove convertit une chaîne comme "e2e4" ou "e7e8q" en Move pseudo-légal
// de la position courante.
func parseMove(s string, pos position.Position) (position.Move, error) {
	if len(s) < 4 || len(s) > 5 {
		return position.Move{}, fmt.Errorf("coup invalide %q — ex: e2e4", s)
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
			return position.Move{}, fmt.Errorf("promotion invalide : '%c' (q/r/b/n)", s[4])
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

// ── Point d'entrée ────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
