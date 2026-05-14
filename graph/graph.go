package graph

import (
	"fmt"
)

// ConditionalEdge represents a conditional branching from a node.
type ConditionalEdge[State any] struct {
	Condition ConditionFunc[State]
	Mapping   map[string]string // Maps the output of Condition to the next node name
}

// StateGraph is a builder for a stateful workflow graph.
type StateGraph[State any] struct {
	nodes            map[string]NodeFunc[State]
	edges            map[string]string
	conditionalEdges map[string]ConditionalEdge[State]
	entryPoint       string
}

// NewStateGraph initializes a new StateGraph builder.
func NewStateGraph[State any]() *StateGraph[State] {
	return &StateGraph[State]{
		nodes:            make(map[string]NodeFunc[State]),
		edges:            make(map[string]string),
		conditionalEdges: make(map[string]ConditionalEdge[State]),
	}
}

// AddNode adds a new node to the graph.
func (g *StateGraph[State]) AddNode(name string, action NodeFunc[State]) *StateGraph[State] {
	if name == START || name == END {
		panic("cannot use START or END as node names")
	}
	g.nodes[name] = action
	return g
}

// AddEdge adds a directed edge from one node to another.
func (g *StateGraph[State]) AddEdge(from, to string) *StateGraph[State] {
	g.edges[from] = to
	return g
}

// AddConditionalEdge adds a conditional edge from a node.
// The condition function determines the next node based on the state.
// The mapping translates the condition's string output to a node name in the graph.
// If mapping is nil, the output of the condition is used directly as the next node name.
func (g *StateGraph[State]) AddConditionalEdge(from string, condition ConditionFunc[State], mapping map[string]string) *StateGraph[State] {
	g.conditionalEdges[from] = ConditionalEdge[State]{
		Condition: condition,
		Mapping:   mapping,
	}
	return g
}

// SetEntryPoint sets the starting node of the graph.
func (g *StateGraph[State]) SetEntryPoint(node string) *StateGraph[State] {
	g.entryPoint = node
	return g
}

// SetFinishPoint sets a default node to END edge.
func (g *StateGraph[State]) SetFinishPoint(node string) *StateGraph[State] {
	return g.AddEdge(node, END)
}

// Compile validates and compiles the StateGraph into an executable Runnable.
func (g *StateGraph[State]) Compile() (*Runnable[State], error) {
	if g.entryPoint == "" {
		return nil, fmt.Errorf("entry point is not set")
	}

	// Basic validation: check if entry point exists
	if _, ok := g.nodes[g.entryPoint]; !ok && g.entryPoint != END {
		return nil, fmt.Errorf("entry point node '%s' does not exist", g.entryPoint)
	}

	return &Runnable[State]{
		nodes:            g.nodes,
		edges:            g.edges,
		conditionalEdges: g.conditionalEdges,
		entryPoint:       g.entryPoint,
	}, nil
}
