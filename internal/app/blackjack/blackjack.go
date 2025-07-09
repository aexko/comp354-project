package blackjack

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// -------------------- ENUM: Suit --------------------

// Enumeration of card suits.
type Suit int // Suit represents a card suit in a standard deck.

const (
	// Heart is the suit of hearts (♥).
	Heart Suit = iota
	// Diamond is the suit of diamonds (♦).
	Diamond
	// Club is the suit of clubs (♣).
	Club
	// Spade is the suit of spades (♠).
	Spade
)

// String returns the string representation of a Suit.
func (suit Suit) String() string {
	return [...]string{"♥", "♦", "♣", "♠"}[suit]
}

// -------------------- STRUCT: Card --------------------

// Representation of a playing card.
type Card struct {
	suit Suit // Suit of the card (Heart, Diamond, Club, Spade).
	rank int  // 1–13 where 1 is Ace, 11 is Jack, 12 is Queen, 13 is King
}

// String returns the string representation of a Card, e.g., "A♥" or "K♠".
func (card Card) String() string {
	rankStr := strconv.Itoa(card.rank)
	switch card.rank {
	case 1:
		rankStr = "A" // Ace
	case 11:
		rankStr = "J" // Jack
	case 12:
		rankStr = "Q" // Queen
	case 13:
		rankStr = "K" // King
	}

	return fmt.Sprintf("%s%s", rankStr, card.suit)
}

// -------------------- STRUCT: Deck --------------------

// Holds a collection of Cards.
type Deck struct {
	cards []Card // Slice of cards in the deck.
}

// Creates a new 52-card deck and randomizes the order of the cards (Shuffle).
func (deck *Deck) Shuffle() {
	deck.cards = make([]Card, 0, 52)
	for suit := Heart; suit <= Spade; suit++ {
		for rank := 1; rank <= 13; rank++ {
			deck.cards = append(deck.cards, Card{suit: suit, rank: rank})
		}
	}
	rand.Shuffle(len(deck.cards), func(i, j int) {
		deck.cards[i], deck.cards[j] = deck.cards[j], deck.cards[i]
	})
}

// Pops a card off the top of the deck.
// If the deck is empty, it returns a negative‐value Card.
func (deck *Deck) Deal() Card {
	if len(deck.cards) == 0 {
		return Card{-1, -1}
	}
	// take from end
	card := deck.cards[len(deck.cards)-1]
	deck.cards = deck.cards[:len(deck.cards)-1]
	return card
}

// -------------------- STRUCT: Player --------------------

// Representation of a blackjack participant (dealer or player).
type Player struct {
	hand  []Card
	score int // Current score based on hand.
	name  string
}

// Draws one card from the deck into the player's hand.
func (player *Player) Hit(deck *Deck) {
	card := deck.Deal()
	player.hand = append(player.hand, card)
	player.UpdateScore()
}

// Returns the player's current score.
func (player *Player) GetScore() int {
	return player.score
}

// Returns a copy of the player's hand.
func (player *Player) GetHand() []Card {
	return player.hand
}

// SetHand sets the Player's hand to the provided slice of cards.
func (player *Player) SetHand(hand []Card) {
	player.hand = hand
}

// Calculates and updates a player's score
// Aces are counted as 11 unless it causes a bust, then as 1; face cards (J, Q, K) are 10.
func (player *Player) UpdateScore() {
	sum := 0
	aces := 0
	for _, c := range player.GetHand() {
		switch c.rank {
		case 11, 12, 13: // Jack, Queen, King
			sum += 10
		case 1:
			sum += 11
			aces++
		default:
			sum += c.rank
		}
	}
	// adjust for aces if bust
	for sum > 21 && aces > 0 {
		sum -= 10
		aces--
	}
	player.score = sum
}

// -------------------- STRUCT: BlackJackGame --------------------

// Representation of a blackjack game.
type BlackJackGame struct {
	deck           *Deck
	dealer         *Player
	players        []*Player
	numberOfRounds int    // Total number of rounds to play.
	currentRound   int    // Current round number.
	currentPlayer  int    // Index of the current player in players slice.
	phase          string // Current game phase ("deal", "player_turn", "dealer_turn", "round_end", "game_over").
}

// NewBlackJackGame creates a new Blackjack game with the specified number of players and rounds.
// The first player is treated as a regular player, and a separate dealer is created.
func NewBlackJackGame(numPlayers, rounds int) *BlackJackGame {
	players := make([]*Player, numPlayers)
	for i := 0; i < numPlayers; i++ {
		//players[i] = &Player{name: strconv.Itoa(i + 1)}
		players[i] = &Player{name: fmt.Sprintf("Player %d", i+1)}
	}
	return &BlackJackGame{
		deck:           &Deck{},
		dealer:         &Player{name: "Dealer"},
		players:        players,
		numberOfRounds: rounds,
		currentRound:   1,
		currentPlayer:  0,
		phase:          "deal",
	}
}

// dealInitialCards deals two cards to each player and the dealer, resetting their hands and scores.
func (game *BlackJackGame) dealInitialCards() {
	for _, player := range game.players {
		player.SetHand(nil)
		player.score = 0
		for i := 0; i < 2; i++ {
			player.Hit(game.deck)
		}
	}
	game.dealer.SetHand(nil)
	game.dealer.score = 0
	for i := 0; i < 2; i++ {
		game.dealer.Hit(game.deck)
	}
}

// model represents the Bubbletea UI model for the Blackjack game.
type model struct {
	game          *BlackJackGame // Game state and logic.
	cardStyle     lipgloss.Style // Style for rendering cards.
	headerStyle   lipgloss.Style // Style for headers and prompts.
	tableStyle    lipgloss.Style // Style for the game table.
	nameStyle     lipgloss.Style // Style for player/dealer names.
	cardAreaStyle lipgloss.Style // Style for card display area.
	scoreStyle    lipgloss.Style // Style for score display.
}

// initialModel creates a new Bubbletea model with a Blackjack game initialized for 2 players and 3 rounds.
func initialModel() tea.Model {
	game := NewBlackJackGame(2, 3)
	return model{
		game:          game,
		cardStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true), // Bright blue, bold text for cards
		headerStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")), // Bold green text for headers
		tableStyle:    lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderTop(true).BorderBottom(true).Width(55),
		nameStyle:     lipgloss.NewStyle().Width(10).Align(lipgloss.Left),
		cardAreaStyle: lipgloss.NewStyle().Width(29).Align(lipgloss.Left),
		scoreStyle:    lipgloss.NewStyle().Width(12).Align(lipgloss.Right),
	}
}

// Init initializes the model by shuffling the deck and returns no initial commands.
func (m model) Init() tea.Cmd {
	m.game.deck.Shuffle()
	return nil
}

// Update handles user input and updates the game state based on the current phase.
// It processes key presses to hit, stand, advance rounds, or quit, and manages phase transitions.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.game.phase {
	case "deal":
		// Deal initial cards and move to player turns
		m.game.dealInitialCards()
		m.game.phase = "player_turn"
	case "player_turn":
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "h":
				// Current player draws a card
				player := m.game.players[m.game.currentPlayer]
				player.Hit(m.game.deck)
				if player.GetScore() > 21 {
					// Player busts, move to next player or dealer
					m.game.currentPlayer++
					if m.game.currentPlayer >= len(m.game.players) {
						m.game.phase = "dealer_turn"
					}
				}
			case "s":
				// Player stands, move to next player or dealer
				m.game.currentPlayer++
				if m.game.currentPlayer >= len(m.game.players) {
					m.game.phase = "dealer_turn"
				}
			}
		}
	case "dealer_turn":
		// Dealer hits until score is at least 17
		for m.game.dealer.GetScore() < 17 {
			m.game.dealer.Hit(m.game.deck)
		}
		m.game.phase = "round_end"
	case "round_end":
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "enter", " ":
				// Advance to next round or end game
				m.game.currentRound++
				m.game.currentPlayer = 0
				if m.game.currentRound > m.game.numberOfRounds {
					m.game.phase = "game_over"
				} else {
					m.game.deck.Shuffle()
					m.game.phase = "deal"
				}
			}
		}
	case "game_over":
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

// View renders the current game state as a string for display in the terminal.
// It shows the dealer's hand, players' hands and scores, and prompts based on the game phase.
func (m model) View() string {
	var rows []string

	// Render header with round information
	rows = append(rows, m.headerStyle.Render(fmt.Sprintf("Blackjack - Round %d/%d", m.game.currentRound, m.game.numberOfRounds)))

	// Render dealer's hand
	dealerHand := m.game.dealer.GetHand()
	var dealerCards string
	if m.game.phase == "deal" || m.game.phase == "player_turn" {
		if len(dealerHand) > 0 {
			dealerCards = m.cardStyle.Render(dealerHand[0].String()) + " [Hidden]"
		} else {
			dealerCards = ""
		}
	} else {
		var cards []string
		for _, card := range dealerHand {
			cards = append(cards, m.cardStyle.Render(card.String()))
		}
		dealerCards = strings.Join(cards, " ")
	}
	dealerScore := ""
	if m.game.phase != "deal" && m.game.phase != "player_turn" {
		dealerScore = fmt.Sprintf("(Score: %d)", m.game.dealer.GetScore())
	}
	dealerRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.nameStyle.Render("Dealer:"),
		m.cardAreaStyle.Render(dealerCards),
		m.scoreStyle.Render(dealerScore),
	)
	rows = append(rows, dealerRow)

	// Render players' hands
	for _, player := range m.game.players {
		var cards []string
		for _, card := range player.GetHand() {
			cards = append(cards, m.cardStyle.Render(card.String()))
		}
		playerRow := lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.nameStyle.Render(player.name+":"),
			m.cardAreaStyle.Render(strings.Join(cards, " ")),
			m.scoreStyle.Render(fmt.Sprintf("(Score: %d)", player.GetScore())),
		)
		rows = append(rows, playerRow)
	}

	// Combine rows with table style
	s := m.tableStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))

	// Render phase-specific prompts
	switch m.game.phase {
	case "player_turn":
		s += fmt.Sprintf("\n%s's turn: Press 'h' to hit, 's' to stand", m.game.players[m.game.currentPlayer].name)
	case "dealer_turn":
		s += "\nDealer's turn"
	case "round_end":
		s += m.headerStyle.Render("\nRound Results:")
		dealerScore := m.game.dealer.GetScore()
		for i, player := range m.game.players {
			playerScore := player.GetScore()
			result := "loses"
			if playerScore <= 21 && (dealerScore > 21 || playerScore > dealerScore) {
				result = "wins"
			}
			s += fmt.Sprintf("\nPlayer %d %s (Player: %d vs Dealer: %d)", i+1, result, playerScore, dealerScore)
		}
		s += "\nPress Enter or Space to continue"
	case "game_over":
		s += m.headerStyle.Render("\nGame Over!")
		s += "\nPress 'q' to quit"
	}

	return s
}

// Run starts the Bubbletea program to run the Blackjack game with its UI.
func Run() {
	program := tea.NewProgram(initialModel())
	if _, err := program.Run(); err != nil {
		panic(err) // Panic on program run error
	}
}
