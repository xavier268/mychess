package main

import (
	"context"
	"fmt"
	"mychess"
	"strings"
	"time"

	"mychess/game"
	"mychess/position"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	-position.KING:   "♔",
	-position.QUEEN:  "♕",
	-position.ROOK:   "♖",
	-position.BISHOP: "♗",
	-position.KNIGHT: "♘",
	-position.PAWN:   "♙",
	position.EMPTY:   " ",
}

var (
	lightSqBg = lipgloss.Color("#4A90D9")
	darkSqBg  = lipgloss.Color("#2E7D32")
	whiteFg   = lipgloss.Color("#FFFFFF")
	blackFg   = lipgloss.Color("#000000")

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
	g   *game.Game
	ctx context.Context

	input     string
	showStats bool // toggle avec 's' ; false = tick arrêté
	gameOver  bool // true quand la partie est terminée (mat ou pat)

	// displayed messages
	history string // history of moves played in the game, updated after each move
	message string // feedback message for the user
	stats   string // stats extracted from the game and updated every tick (500ms)

	cancel context.CancelFunc // to stop everything when user wishes to quit

	// displayPos est une copie stable de la position, mise à jour uniquement
	// lorsqu'un coup est joué ou annulé — jamais lors du tick.
	// Cela évite que l'échiquier clignote pendant l'analyse (AlphaBeta
	// modifie g.Position en continu via DoMove/UndoMove).
	displayPos position.Position
}

func initialModel() model {
	ctx, cancel := context.WithCancel(context.Background())
	g := game.NewGame()
	m := model{g: g, ctx: ctx, cancel: cancel, displayPos: g.Position}
	g.AnalysisAsync(ctx, analysisDepth)
	return m
}

func (m model) Init() tea.Cmd {
	return nil // showStats démarre à false, le tick est lancé uniquement sur demande
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

// ── Helpers d'affichage ───────────────────────────────────────────────────────

// moveStr converts a Move to its string representation (e.g. "e2e4", "e7e8q", "e1g1").
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

// buildHistory formats the game history as "1. e2e4 e7e5\n2. ..."
// Only the last maxLines lines are shown; a "…" prefix indicates truncation.
func buildHistory(history []position.Move, maxLines int) string {
	// Each line holds one move pair (white + optional black).
	// Compute the first half-move index to display.
	totalPairs := (len(history) + 1) / 2
	firstPair := 0
	truncated := false
	if totalPairs > maxLines {
		firstPair = totalPairs - maxLines
		truncated = true
	}
	firstHalf := firstPair * 2

	var sb strings.Builder
	if truncated {
		sb.WriteString("…\n")
	}
	for i, mv := range history[firstHalf:] {
		idx := firstHalf + i
		if idx%2 == 0 {
			fmt.Fprintf(&sb, "%d. ", idx/2+1)
		}
		sb.WriteString(moveStr(mv))
		if idx%2 == 0 {
			sb.WriteString(" ")
		} else {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// buildStats reads analysis results from the game engine (no lock needed — display only).
func buildStats(g *game.Game) string {
	entry := g.LastRootEntry
	best := "–"
	if entry.Best.From != entry.Best.To || entry.Best.Promotion != position.EMPTY {
		best = moveStr(entry.Best)
	}
	return fmt.Sprintf("Profondeur: %d | Score: %+d | Meilleur: %s\n%s",
		entry.Depth, entry.Score, best, g.Z.Stats())
}

// renderBoard draws the 8×8 board for the given position.
func renderBoard(pos position.Position) string {
	var sb strings.Builder
	sb.WriteString("  a  b  c  d  e  f  g  h\n")
	for rank := 7; rank >= 0; rank-- {
		fmt.Fprintf(&sb, "%d", rank+1)
		for file := 0; file < 8; file++ {
			sq := position.Sq(rank, file)
			piece := pos.PieceAt(sq)
			sym := unicodePiece[piece]
			isDark := (rank+file)%2 == 0
			bg := lightSqBg
			if isDark {
				bg = darkSqBg
			}
			fg := whiteFg
			if piece < 0 {
				fg = blackFg
			}
			cell := lipgloss.NewStyle().Background(bg).Foreground(fg).Render(" " + sym + " ")
			sb.WriteString(cell)
		}
		fmt.Fprintf(&sb, " %d\n", rank+1)
	}
	sb.WriteString("  a  b  c  d  e  f  g  h")
	return sb.String()
}

// ── Fin de partie ─────────────────────────────────────────────────────────────

// hasLegalMove retourne true s'il existe au moins un coup légal dans la position courante.
// GetMoveList retourne des coups pseudo-légaux ; on filtre l'illégalité exactement
// comme AlphaBeta : un coup est illégal s'il laisse le roi du joueur en prise.
func (m *model) hasLegalMove() bool {
	p := m.g.Position
	mover := p.Turn()
	for _, mv := range p.GetMoveList() {
		newPos, _ := p.DoMove(mv)
		if !newPos.IsSquareAttacked(newPos.KingPosition(mover), 1^mover) {
			return true
		}
	}
	return false
}

// checkGameOver détecte mat et pat et retourne true si la partie est terminée.
// Doit être appelé APRES chaque coup, avant de relancer l'analyse.
func (m *model) checkGameOver() bool {
	if m.hasLegalMove() {
		return m.checkRepeat3()
	}
	m.gameOver = true
	if m.g.Position.IsCheck() {
		winner := "Noirs"
		if m.g.Position.Turn() != position.WHITE {
			winner = "Blancs"
		}
		m.message = errStyle.Render(fmt.Sprintf("ÉCHEC ET MAT — Les %s gagnent !", winner))
	} else {
		m.message = boldStyle.Render("PAT — Partie nulle !")
	}
	return true
}

// Verifie si dans la liste des coups historiques, on a répété 3 fois la même position
func (m *model) checkRepeat3() bool {
	pm := make(map[uint64]int, len(m.g.History)+1)
	for _, move := range m.g.History {
		pos := move.PrevHash
		pm[pos]++
		if pm[pos] >= 3 {
			m.gameOver = true
			m.message = boldStyle.Render("PAT par 3 répétition de positions identiques — Partie nulle !")
			return true
		}
	}
	return false
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "x":
			m.cancel()
			return m, tea.Quit
		}

		if m.gameOver {
			return m, nil
		}

		switch msg.String() {
		case "s":
			m.showStats = !m.showStats
			if m.showStats {
				m.stats = buildStats(m.g)
				return m, tick() // démarre la boucle de tick
			}
			return m, nil // la boucle s'arrêtera au prochain tickMsg
		case "a":
			err := m.g.AutoPlay()
			if err != nil {
				m.message = errStyle.Render("autoplay: " + err.Error())
			} else {
				m.displayPos = m.g.Position
				m.history = buildHistory(m.g.History, 22)
				if !m.checkGameOver() {
					m.message = okStyle.Render("coup automatique joué")
					m.g.AnalysisAsync(m.ctx, analysisDepth)
				}
			}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}

		case "enter":
			input := strings.TrimSpace(m.input)
			m.input = ""
			switch input {
			case "":
				// ignore empty input
			default:
				mv, err := parseMove(input, m.displayPos)
				if err != nil {
					m.message = errStyle.Render(err.Error())
				} else {
					m.g.Play(mv)
					m.displayPos = m.g.Position
					m.history = buildHistory(m.g.History, 22)
					if !m.checkGameOver() {
						m.message = okStyle.Render("joué : " + input)
						m.g.AnalysisAsync(m.ctx, analysisDepth)
					}
				}
			}

		default:
			if k := msg.Key(); k.Text != "" {
				m.input += k.Text
			}
		}

	case tickMsg:
		if !m.showStats {
			return m, nil // boucle arrêtée
		}
		m.stats = buildStats(m.g)
		return m, tick()
	}
	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m model) View() tea.View {
	turn := "Blancs"
	if m.displayPos.Turn() == position.BLACK {
		turn = "Noirs"
	}

	// Colonne gauche : trait + échiquier + analyse + saisie
	var left strings.Builder
	score := m.g.LastRootEntry.Score
	left.WriteString("\n" + mychess.COPYRIGHT + " V" + mychess.VERSION + "\n(build " + mychess.BUILDDATE + " - " + mychess.BUILDHASH + ")\n")
	if m.gameOver {
		left.WriteString(boldStyle.Render("\n══════ PARTIE TERMINÉE ══════") + "\n\n")
	} else {
		left.WriteString(boldStyle.Render("\nTrait aux "+turn) + fmt.Sprintf("  (score : %+d)", score) + "\n\n")
	}

	left.WriteString(renderBoard(m.displayPos) + "\n\n")
	if m.showStats {
		left.WriteString(boldStyle.Render("Analyse :") + "\n")
		if m.stats != "" {
			left.WriteString(infoStyle.Render(m.stats) + "\n")
		}
		left.WriteString("\n")
	}
	if !m.gameOver {
		left.WriteString(boldStyle.Render("Coup :") + " " + m.input + "_\n")
	}
	if m.message != "" {
		left.WriteString(m.message + "\n")
	}
	if m.gameOver {
		left.WriteString("\n" + infoStyle.Render("[x=quitter]"))
	} else {
		left.WriteString("\n" + infoStyle.Render("[entrer=jouer  a=autoPlay  s=analyse  x=quitter]"))
	}

	// Colonne droite : historique
	var right strings.Builder
	right.WriteString(boldStyle.Render("\n\n\n\nHistorique :") + "\n\n")
	if m.history != "" {
		right.WriteString(m.history)
	} else {
		right.WriteString("–\n")
	}

	cols := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().MarginRight(4).Render(left.String()),
		right.String(),
	)
	return tea.NewView(cols)
}

// ── Point d'entrée ────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Erreur fatale : %v\n", err)
	}
}
