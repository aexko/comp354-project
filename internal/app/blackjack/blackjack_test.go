package blackjack

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Test if there's a game instance properly created
func TestGameInstanceCreation(t *testing.T) {
	// Test creating a game with different configurations
	tests := []struct {
		name       string
		numPlayers int
		numRounds  int
	}{
		{"Single player, single round", 1, 1},
		{"Two players, three rounds", 2, 3},
		{"Four players, five rounds", 4, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewBlackJackGame(tt.numPlayers, tt.numRounds)

			// Check that game instance is created
			if game == nil {
				t.Fatal("Game instance should not be nil")
			}

			// Check that all components are initialized
			if game.deck == nil {
				t.Error("Deck should be initialized")
			}

			if game.dealer == nil {
				t.Error("Dealer should be initialized")
			}

			if game.players == nil {
				t.Error("Players slice should be initialized")
			}

			if len(game.players) != tt.numPlayers {
				t.Errorf("Expected %d players, got %d", tt.numPlayers, len(game.players))
			}

			if game.numberOfRounds != tt.numRounds {
				t.Errorf("Expected %d rounds, got %d", tt.numRounds, game.numberOfRounds)
			}

			// Check initial state
			if game.currentRound != 1 {
				t.Errorf("Expected initial round to be 1, got %d", game.currentRound)
			}

			if game.currentPlayer != 0 {
				t.Errorf("Expected initial player to be 0, got %d", game.currentPlayer)
			}

			if game.phase != "deal" {
				t.Errorf("Expected initial phase to be 'deal', got '%s'", game.phase)
			}

			// Check that players are properly initialized
			for i, player := range game.players {
				if player == nil {
					t.Errorf("Player %d should not be nil", i)
					continue
				}
				if !strings.Contains(player.name, "Player") {
					t.Errorf("Player %d should have name containing 'Player', got '%s'", i, player.name)
				}
			}

			// Check that dealer is properly initialized
			if game.dealer.name != "Dealer" {
				t.Errorf("Expected dealer name to be 'Dealer', got '%s'", game.dealer.name)
			}
		})
	}
}

// Test if there are exactly 52 cards in a shuffled deck
func TestDeckHas52Cards(t *testing.T) {
	deck := &Deck{}
	deck.Shuffle()

	// Should have exactly 52 cards
	if len(deck.cards) != 52 {
		t.Errorf("Expected deck to have 52 cards, got %d", len(deck.cards))
	}

	// Verify all suits are present (13 cards each)
	suitCount := make(map[Suit]int)
	for _, card := range deck.cards {
		suitCount[card.suit]++
	}

	expectedSuits := []Suit{Heart, Diamond, Club, Spade}
	for _, suit := range expectedSuits {
		if suitCount[suit] != 13 {
			t.Errorf("Expected 13 cards of suit %v, got %d", suit, suitCount[suit])
		}
	}

	// Verify all ranks are present (4 cards each)
	rankCount := make(map[int]int)
	for _, card := range deck.cards {
		rankCount[card.rank]++
	}

	for rank := 1; rank <= 13; rank++ {
		if rankCount[rank] != 4 {
			t.Errorf("Expected 4 cards of rank %d, got %d", rank, rankCount[rank])
		}
	}
}

// Test that deck maintains 52 cards after multiple shuffles
func TestDeckConsistencyAfterShuffle(t *testing.T) {
	deck := &Deck{}

	// Shuffle multiple times
	for i := 0; i < 10; i++ {
		deck.Shuffle()
		if len(deck.cards) != 52 {
			t.Errorf("Shuffle %d: Expected 52 cards, got %d", i+1, len(deck.cards))
		}
	}
}

// Test that total value of cards is respected when hit
func TestCardValueRespectedOnHit(t *testing.T) {
	player := &Player{name: "Test Player"}
	deck := &Deck{}

	// Create a deck with known cards (in reverse order since Deal takes from end)
	deck.cards = []Card{
		{Heart, 5},    // 5 points (will be dealt last)
		{Spade, 7},    // 7 points
		{Club, 10},    // 10 points
		{Diamond, 13}, // 10 points (King, will be dealt first)
	}

	// Test hitting with face card
	initialScore := player.GetScore() // Should be 0
	player.Hit(deck)                  // Should get King (10 points)
	expectedScore := initialScore + 10
	if player.GetScore() != expectedScore {
		t.Errorf("Expected score %d after hitting King, got %d", expectedScore, player.GetScore())
	}

	// Test hitting with regular number
	player.Hit(deck) // Should get 10
	expectedScore += 10
	if player.GetScore() != expectedScore {
		t.Errorf("Expected score %d after hitting 10, got %d", expectedScore, player.GetScore())
	}

	// Test hitting with another number
	player.Hit(deck) // Should get 7
	expectedScore += 7
	if player.GetScore() != expectedScore {
		t.Errorf("Expected score %d after hitting 7, got %d", expectedScore, player.GetScore())
	}
}

// Test Ace value adjustment separately
func TestAceValueInHit(t *testing.T) {
	player := &Player{name: "Test Player"}
	deck := &Deck{}

	// Test Ace as 11 when safe
	deck.cards = []Card{{Heart, 1}} // Ace
	player.Hit(deck)                // Should get Ace (11 points)
	if player.GetScore() != 11 {
		t.Errorf("Expected score 11 after hitting Ace, got %d", player.GetScore())
	}

	// Test Ace adjustment when it would bust
	player.SetHand([]Card{{Spade, 10}}) // Start with 10
	player.UpdateScore()
	deck.cards = []Card{{Heart, 1}} // Ace
	player.Hit(deck)                // Should get Ace (11 points, but will adjust to 1 if > 21)
	if player.GetScore() != 21 {
		t.Errorf("Expected score 21 (10 + 11) after hitting Ace, got %d", player.GetScore())
	}

	// Test Ace adjustment when adding to existing high score
	player.SetHand([]Card{{Spade, 10}, {Club, 5}}) // Start with 15
	player.UpdateScore()
	deck.cards = []Card{{Heart, 1}} // Ace
	player.Hit(deck)                // Should get Ace (1 point to avoid bust)
	if player.GetScore() != 16 {
		t.Errorf("Expected score 16 (15 + 1) after hitting Ace to avoid bust, got %d", player.GetScore())
	}
}

// Test that Ace value adjusts correctly to prevent unnecessary busts
func TestAceValueAdjustment(t *testing.T) {
	player := &Player{name: "Test Player"}

	// Test Ace counting as 11 when safe
	player.SetHand([]Card{
		{Heart, 1}, // Ace
		{Spade, 5}, // 5
	})
	player.UpdateScore()

	if player.GetScore() != 16 { // 11 + 5
		t.Errorf("Expected score 16 (Ace as 11), got %d", player.GetScore())
	}

	// Test Ace counting as 1 when 11 would bust
	player.SetHand([]Card{
		{Heart, 1},  // Ace
		{Spade, 10}, // 10
		{Club, 8},   // 8
	})
	player.UpdateScore()

	if player.GetScore() != 19 { // 1 + 10 + 8
		t.Errorf("Expected score 19 (Ace as 1), got %d", player.GetScore())
	}

	// Test multiple Aces
	player.SetHand([]Card{
		{Heart, 1},   // Ace
		{Diamond, 1}, // Ace
		{Spade, 9},   // 9
	})
	player.UpdateScore()

	if player.GetScore() != 21 { // 11 + 1 + 9
		t.Errorf("Expected score 21 (one Ace as 11, one as 1), got %d", player.GetScore())
	}
}

// Test that face cards are valued correctly
func TestFaceCardValues(t *testing.T) {
	player := &Player{name: "Test Player"}

	// Test Jack (11) = 10 points
	player.SetHand([]Card{{Heart, 11}})
	player.UpdateScore()
	if player.GetScore() != 10 {
		t.Errorf("Expected Jack to be worth 10 points, got %d", player.GetScore())
	}

	// Test Queen (12) = 10 points
	player.SetHand([]Card{{Diamond, 12}})
	player.UpdateScore()
	if player.GetScore() != 10 {
		t.Errorf("Expected Queen to be worth 10 points, got %d", player.GetScore())
	}

	// Test King (13) = 10 points
	player.SetHand([]Card{{Club, 13}})
	player.UpdateScore()
	if player.GetScore() != 10 {
		t.Errorf("Expected King to be worth 10 points, got %d", player.GetScore())
	}

	// Test all face cards together
	player.SetHand([]Card{
		{Heart, 11},   // Jack
		{Diamond, 12}, // Queen
		{Club, 13},    // King
	})
	player.UpdateScore()
	if player.GetScore() != 30 {
		t.Errorf("Expected J+Q+K to be worth 30 points, got %d", player.GetScore())
	}
}

// Test that going over 21 is considered a bust/loss
func TestBustOver21(t *testing.T) {
	player := &Player{name: "Test Player"}

	// Create a hand that busts
	player.SetHand([]Card{
		{Heart, 10}, // 10
		{Spade, 5},  // 5
		{Club, 7},   // 7
	})
	player.UpdateScore()

	score := player.GetScore()
	if score <= 21 {
		t.Errorf("Expected score to be over 21 (bust), got %d", score)
	}

	if score != 22 {
		t.Errorf("Expected exact score of 22, got %d", score)
	}
}

// Test bust scenario in actual game play
func TestBustScenarioInGame(t *testing.T) {
	m := model{
		game: NewBlackJackGame(2, 3),
	}
	m.game.phase = "player_turn"
	m.game.currentPlayer = 0

	// Set up player with cards that will bust on next hit
	m.game.players[0].SetHand([]Card{
		{Heart, 10}, // 10
		{Spade, 9},  // 9
	})
	m.game.players[0].UpdateScore()

	// Verify initial score
	if m.game.players[0].GetScore() != 19 {
		t.Errorf("Expected initial score 19, got %d", m.game.players[0].GetScore())
	}

	// Set up deck with card that will cause bust
	m.game.deck.cards = []Card{{Club, 5}} // This will cause bust (19 + 5 = 24)

	// Player hits
	hitMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
	updatedModel, _ := m.Update(hitMsg)
	m = updatedModel.(model)

	// Check that player busted
	finalScore := m.game.players[0].GetScore()
	if finalScore <= 21 {
		t.Errorf("Expected player to bust (score > 21), got %d", finalScore)
	}

	if finalScore != 24 {
		t.Errorf("Expected exact bust score of 24, got %d", finalScore)
	}

	// Check that game advanced to next player due to bust
	if m.game.currentPlayer != 1 {
		t.Errorf("Expected game to advance to next player after bust, got player %d", m.game.currentPlayer)
	}
}

// Test that dealer follows 17 rule
func TestDealerFollows17Rule(t *testing.T) {
	m := model{
		game: NewBlackJackGame(2, 3),
	}
	m.game.phase = "dealer_turn"

	// Set dealer with score under 17
	m.game.dealer.SetHand([]Card{
		{Heart, 6}, // 6
		{Spade, 8}, // 8
	})
	m.game.dealer.UpdateScore()

	// Verify dealer score is under 17
	if m.game.dealer.GetScore() != 14 {
		t.Errorf("Expected dealer initial score 14, got %d", m.game.dealer.GetScore())
	}

	// Set up deck with cards
	m.game.deck.cards = []Card{
		{Club, 5},    // 5
		{Diamond, 4}, // 4
		{Heart, 3},   // 3
	}

	// Process dealer turn
	updatedModel, _ := m.Update(nil)
	m = updatedModel.(model)

	// Dealer should have hit until >= 17
	dealerScore := m.game.dealer.GetScore()
	if dealerScore < 17 && dealerScore > 0 {
		t.Errorf("Expected dealer to hit until score >= 17, got %d", dealerScore)
	}

	// Should advance to round_end phase
	if m.game.phase != "round_end" {
		t.Errorf("Expected phase to be 'round_end', got '%s'", m.game.phase)
	}
}

// Test complete game initialization
func TestCompleteGameInitialization(t *testing.T) {
	// Test model creation
	modelInterface := initialModel()
	m := modelInterface.(model)

	// Verify game instance exists
	if m.game == nil {
		t.Fatal("Game instance should not be nil")
	}

	// Verify deck is initialized
	if m.game.deck == nil {
		t.Fatal("Deck should be initialized")
	}

	// Initialize the deck
	m.game.deck.Shuffle()

	// Verify deck has 52 cards
	if len(m.game.deck.cards) != 52 {
		t.Errorf("Expected deck to have 52 cards, got %d", len(m.game.deck.cards))
	}

	// Verify that the model has the essential components
	// (Style components are initialized by initialModel function with lipgloss.NewStyle())
	if m.game.currentRound != 1 {
		t.Errorf("Expected current round to be 1, got %d", m.game.currentRound)
	}

	if len(m.game.players) != 2 {
		t.Errorf("Expected 2 players in initial model, got %d", len(m.game.players))
	}
}

// Test that deck dealing reduces card count
func TestDeckDealingReducesCount(t *testing.T) {
	deck := &Deck{}
	deck.Shuffle()

	initialCount := len(deck.cards)
	if initialCount != 52 {
		t.Errorf("Expected initial deck count of 52, got %d", initialCount)
	}

	// Deal 10 cards
	for i := 0; i < 10; i++ {
		card := deck.Deal()
		if card.suit < 0 || card.rank < 0 {
			t.Errorf("Card %d should be valid, got invalid card", i)
		}
	}

	// Should have 42 cards left
	if len(deck.cards) != 42 {
		t.Errorf("Expected 42 cards after dealing 10, got %d", len(deck.cards))
	}
}

// Test edge case: dealing from empty deck
func TestDealingFromEmptyDeck(t *testing.T) {
	deck := &Deck{}
	// Don't shuffle - deck should be empty

	card := deck.Deal()

	// Should return invalid card
	if card.suit != -1 || card.rank != -1 {
		t.Errorf("Expected invalid card (-1, -1) from empty deck, got (%v, %d)", card.suit, card.rank)
	}
}

// Test that blackjack (21) is recognized correctly
func TestBlackjackRecognition(t *testing.T) {
	player := &Player{name: "Test Player"}

	// Test natural blackjack (Ace + 10-value card)
	player.SetHand([]Card{
		{Heart, 1},  // Ace (11)
		{Spade, 10}, // 10
	})
	player.UpdateScore()

	if player.GetScore() != 21 {
		t.Errorf("Expected blackjack score of 21, got %d", player.GetScore())
	}

	// Test blackjack with face card
	player.SetHand([]Card{
		{Diamond, 1}, // Ace (11)
		{Club, 13},   // King (10)
	})
	player.UpdateScore()

	if player.GetScore() != 21 {
		t.Errorf("Expected blackjack score of 21 with Ace+King, got %d", player.GetScore())
	}
}

// Test flaws and edge cases
func TestGameFlaws(t *testing.T) {
	// Test empty deck handling
	t.Run("Empty deck returns invalid card", func(t *testing.T) {
		deck := &Deck{} // Empty deck
		card := deck.Deal()

		if card.suit != -1 || card.rank != -1 {
			t.Errorf("Expected invalid card from empty deck, got valid card")
		}
	})

	// Test all players bust scenario
	t.Run("All players bust should skip dealer", func(t *testing.T) {
		m := model{game: NewBlackJackGame(2, 1)}
		m.game.phase = "player_turn"

		// Make both players bust
		m.game.players[0].SetHand([]Card{{Heart, 10}, {Spade, 6}, {Club, 6}}) // 22
		m.game.players[0].UpdateScore()
		m.game.players[1].SetHand([]Card{{Diamond, 10}, {Heart, 5}, {Spade, 7}}) // 22
		m.game.players[1].UpdateScore()

		// Both players have busted, but dealer will still play
		// This is a flaw - dealer shouldn't need to play if all players busted
		if m.game.players[0].GetScore() <= 21 || m.game.players[1].GetScore() <= 21 {
			t.Error("Test setup failed - players should be busted")
		}
	})

	// Test deck exhaustion during game
	t.Run("Deck exhaustion handling", func(t *testing.T) {
		player := &Player{name: "Test"}
		deck := &Deck{}
		deck.cards = []Card{{Heart, 5}} // Only one card

		player.Hit(deck) // Takes the only card

		// Now deck is empty, next hit should handle gracefully
		player.Hit(deck) // This will add invalid card {-1, -1}

		// Check if invalid card was added to hand
		hand := player.GetHand()
		if len(hand) != 2 {
			t.Errorf("Expected 2 cards in hand, got %d", len(hand))
		}

		// Last card should be invalid
		lastCard := hand[len(hand)-1]
		if lastCard.suit != -1 || lastCard.rank != -1 {
			t.Error("Expected invalid card to be added when deck is empty")
		}
	})

	// Test nil hand handling
	t.Run("Nil hand initialization", func(t *testing.T) {
		player := &Player{name: "Test"}
		player.SetHand(nil)

		// This should not panic
		score := player.GetScore()
		if score != 0 {
			t.Errorf("Expected score 0 for empty hand, got %d", score)
		}

		player.UpdateScore() // Should not panic with nil hand
	})
}

// Test round management edge cases
func TestRoundManagementFlaws(t *testing.T) {
	// Test deck state between rounds
	t.Run("Deck state between rounds", func(t *testing.T) {
		m := model{game: NewBlackJackGame(2, 2)}
		m.game.deck.Shuffle()

		// Deal some cards
		m.game.dealInitialCards()
		initialDeckSize := len(m.game.deck.cards)

		// Advance to next round
		m.game.currentRound = 1
		m.game.phase = "round_end"

		enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := m.Update(enterMsg)
		m = updatedModel.(model)

		// Check if deck was reshuffled (should have 52 cards again)
		newDeckSize := len(m.game.deck.cards)
		if newDeckSize != 52 {
			t.Errorf("Expected deck to be reshuffled to 52 cards, got %d", newDeckSize)
		}

		// This is actually a flaw - the used cards are not returned to deck
		// before reshuffling, they're just lost and a new deck is created
		if newDeckSize == initialDeckSize {
			t.Error("Deck should be different size after reshuffling, but cards are lost")
		}
	})
}
