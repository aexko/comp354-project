package blackjack

import (
	"fmt"
	"math/rand"
	"strconv"
)

// -------------------- ENUM: Suit --------------------

// Enumeration of card suits.
type Suit int

const (
	Heart Suit = iota
	Diamond
	Club
	Spade
)

func (suit Suit) String() string {
	return [...]string{"Heart", "Diamond", "Club", "Spade"}[suit]
}

// -------------------- STRUCT: Card --------------------

// Representation of a playing card.
type Card struct {
	suit Suit // Heart-Spade
	rank int  // 1–13
}

func (card Card) String() string {
	return fmt.Sprintf("%s-%d", card.suit, card.rank)
}

// -------------------- STRUCT: Deck --------------------

// Holds a collection of Cards.
type Deck struct {
	cards []Card
}

// Creates a new 52-card deck and randomizes the order of the cards.
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
	score int
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

func (player *Player) SetHand(hand []Card) {
	player.hand = hand
}

// Updates a player's score
func (player *Player) UpdateScore() {
	sum := 0
	aces := 0
	for _, c := range player.GetHand() {
		switch c.rank {
		case 11, 12, 13:
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
	numberOfRounds int
}

// Builds a game with non‐dealer players and R rounds.
func NewBlackJackGame(numPlayers, rounds int) *BlackJackGame {
	players := make([]*Player, numPlayers)
	for i := 0; i < numPlayers; i++ {
		players[i] = &Player{name: strconv.Itoa(i + 1)}
	}
	return &BlackJackGame{
		deck:           &Deck{},
		dealer:         players[0],
		players:        players,
		numberOfRounds: rounds,
	}
}

// Initializes the deck, shuffles, and runs the main loop.
func (game *BlackJackGame) StartGame() {
	fmt.Println("Starting Blackjack...")
	game.deck.Shuffle()
	game.gameLoop()
	game.endGame()
}

// Gives two cards to each player and the dealer.
func (game *BlackJackGame) dealInitialCards() {
	// each player gets 2 cards
	for _, player := range game.players {
		player.hand = nil
		player.score = 0
		for i := 0; i < 2; i++ {
			player.Hit(game.deck)
		}
		fmt.Printf("Player %s's Hand: %s\n\n", player.name, player.hand)
	}
	fmt.Printf("Dealer Hand: %s\n\n", game.dealer.GetHand())
}

// Runs each round: deal, let everyone play, then settle.
func (game *BlackJackGame) gameLoop() {
	for round := 1; round <= game.numberOfRounds; round++ {
		fmt.Printf("\n-- Round %d --\n", round)
		// reset deck & player hands/scores
		game.deck.Shuffle()
		game.dealInitialCards()
		for _, player := range game.players {
			if player == game.dealer {
				continue
			}
			fmt.Printf("Player %s's turn\n", player.name)
			game.hitOrStandConfirmation(player)
		}

		// Dealer's turn (simple rule: hit until 17+)
		fmt.Println("Dealer's turn")
		for game.dealer.GetScore() < 17 {
			game.dealer.Hit(game.deck)
		}
		game.roundEnd()
	}
}

// Asks the player whether to hit or stand.
func (game *BlackJackGame) hitOrStandConfirmation(player *Player) {
	var choice string
	for {
		fmt.Printf("Hand: %v (Score: %d). Hit or Stand? [hit/stand]: ",
			player.GetHand(), player.GetScore())
		fmt.Scanln(&choice)
		if choice == "hit" {
			player.Hit(game.deck)
			if player.GetScore() > 21 {
				fmt.Println("Busted!")
				break
			}
		} else if choice == "stand" {
			break
		}
	}
}

// Compares dealer and players; declares winners.
func (game *BlackJackGame) roundEnd() {
	dealerScore := game.dealer.GetScore()
	fmt.Printf("Dealer's final hand: %v (Score: %d)\n",
		game.dealer.GetHand(), dealerScore)

	for i, player := range game.players {
		playerScore := player.GetScore()
		result := "loses"
		if playerScore <= 21 && (dealerScore > 21 || playerScore > dealerScore) {
			result = "wins"
		}
		fmt.Printf("Player %d %s (Player: %d vs Dealer: %d)\n",
			i+1, result, playerScore, dealerScore)
	}
}

// Cleans up and shows overall stats (not implemented).
func (game *BlackJackGame) endGame() {
	fmt.Println("\nGame over.")
}

func Run() {
	game := NewBlackJackGame(2, 3) // e.g. 2 players, 3 rounds
	game.StartGame()
}

// func Run() {
// 	program := tea.NewProgram(initialModel())

// 	if _, err := program.Run(); err != nil {
// 		panic(err)
// 	}
// }
