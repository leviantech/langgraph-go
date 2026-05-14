package main

import (
	"context"
	"fmt"
	"log"

	"github.com/leviantech/langgraph-go/graph"
)

type MyState struct {
	Messages []string
}

func NodeA(ctx context.Context, state MyState) (MyState, error) {
	fmt.Println("Executing Node A")
	state.Messages = append(state.Messages, "Message from A")
	return state, nil
}

func NodeB(ctx context.Context, state MyState) (MyState, error) {
	fmt.Println("Executing Node B")
	state.Messages = append(state.Messages, "Message from B")
	return state, nil
}

func main() {
	g := graph.NewStateGraph[MyState]()

	g.AddNode("A", NodeA)
	g.AddNode("B", NodeB)

	g.AddEdge("A", "B")
	g.SetFinishPoint("B") // Equivalent to AddEdge("B", graph.END)
	g.SetEntryPoint("A")

	app, err := g.Compile()
	if err != nil {
		log.Fatalf("Failed to compile graph: %v", err)
	}

	initialState := MyState{Messages: []string{}}
	finalState, err := app.Invoke(context.Background(), initialState)
	if err != nil {
		log.Fatalf("Failed to invoke graph: %v", err)
	}

	fmt.Printf("Final state: %+v\n", finalState)
}
