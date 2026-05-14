package main

import (
	"context"
	"fmt"
	"log"

	"github.com/leviantech/langgraph-go/graph"
)

type GameState struct {
	Score int
	Turn  int
}

func RollDice(ctx context.Context, state GameState) (GameState, error) {
	fmt.Printf("Turn %d: Rolling dice...\n", state.Turn+1)
	state.Score += 4 // Simulate rolling 4
	state.Turn++
	return state, nil
}

func Win(ctx context.Context, state GameState) (GameState, error) {
	fmt.Println("Node: Win! You've reached the target score.")
	return state, nil
}

func Lose(ctx context.Context, state GameState) (GameState, error) {
	fmt.Println("Node: Lose! Out of turns.")
	return state, nil
}

func CheckScore(ctx context.Context, state GameState) (string, error) {
	if state.Score >= 10 {
		return "win", nil
	}
	if state.Turn >= 3 {
		return "lose", nil
	}
	return "roll_again", nil
}

func main() {
	g := graph.NewStateGraph[GameState]()

	g.AddNode("Roll", RollDice)
	g.AddNode("Win", Win)
	g.AddNode("Lose", Lose)

	g.AddConditionalEdge("Roll", CheckScore, map[string]string{
		"win":        "Win",
		"lose":       "Lose",
		"roll_again": "Roll",
	})

	g.SetFinishPoint("Win")
	g.SetFinishPoint("Lose")
	g.SetEntryPoint("Roll")

	app, err := g.Compile()
	if err != nil {
		log.Fatalf("Failed to compile graph: %v", err)
	}

	initialState := GameState{Score: 0, Turn: 0}
	
	fmt.Println("--- Game Start ---")
	finalState, err := app.Invoke(context.Background(), initialState)
	if err != nil {
		log.Fatalf("Error running graph: %v", err)
	}
	fmt.Printf("Game Over. Final Score: %d in %d turns\n", finalState.Score, finalState.Turn)
}
