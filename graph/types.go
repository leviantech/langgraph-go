package graph

import "context"

// NodeFunc is the type for the function that will be executed at each node.
// It receives a context and the current state, and returns the modified state and an error.
type NodeFunc[State any] func(ctx context.Context, state State) (State, error)

// ConditionFunc is the type for the function that decides the next edge to take.
// It returns a string representing the name of the next node or edge, and an error.
type ConditionFunc[State any] func(ctx context.Context, state State) (string, error)

const (
	// START is the special constant representing the entry point of the graph.
	START = "__start__"

	// END is the special constant representing the terminal point of the graph.
	END = "__end__"
)
