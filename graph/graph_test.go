package graph_test

import (
	"context"
	"testing"

	"github.com/leviantech/langgraph-go/graph"
)

type TestState struct {
	Count int
}

func NodeA(ctx context.Context, state TestState) (TestState, error) {
	state.Count += 1
	return state, nil
}

func NodeB(ctx context.Context, state TestState) (TestState, error) {
	state.Count += 2
	return state, nil
}

func TestSequentialGraph(t *testing.T) {
	g := graph.NewStateGraph[TestState]()
	
	g.AddNode("A", NodeA)
	g.AddNode("B", NodeB)

	g.AddEdge("A", "B")
	g.AddEdge("B", graph.END)

	g.SetEntryPoint("A")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	initialState := TestState{Count: 0}
	finalState, err := runnable.Invoke(context.Background(), initialState)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if finalState.Count != 3 {
		t.Errorf("Expected count to be 3, got %d", finalState.Count)
	}
}

func TestConditionalGraph(t *testing.T) {
	g := graph.NewStateGraph[TestState]()

	g.AddNode("Start", NodeA) // +1
	g.AddNode("AddOne", NodeA) // +1
	g.AddNode("AddTwo", NodeB) // +2

	// Condition: if count is 1, go to AddOne, else go to AddTwo
	condition := func(ctx context.Context, state TestState) (string, error) {
		if state.Count == 1 {
			return "add_one", nil
		}
		return "add_two", nil
	}

	g.AddConditionalEdge("Start", condition, map[string]string{
		"add_one": "AddOne",
		"add_two": "AddTwo",
	})

	g.SetFinishPoint("AddOne")
	g.SetFinishPoint("AddTwo")
	g.SetEntryPoint("Start")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Test case 1: Count becomes 1 at Start -> goes to AddOne -> final 2
	state1 := TestState{Count: 0}
	final1, err := runnable.Invoke(context.Background(), state1)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
	if final1.Count != 2 {
		t.Errorf("Expected count to be 2, got %d", final1.Count)
	}

	// Test case 2: Start with count 1 -> becomes 2 at Start -> goes to AddTwo -> final 4
	state2 := TestState{Count: 1}
	final2, err := runnable.Invoke(context.Background(), state2)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
	if final2.Count != 4 {
		t.Errorf("Expected count to be 4, got %d", final2.Count)
	}
}
