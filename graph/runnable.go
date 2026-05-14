package graph

import (
	"context"
	"fmt"
	"strings"
)

// Runnable represents a compiled, executable state graph.
type Runnable[State any] struct {
	nodes            map[string]NodeFunc[State]
	edges            map[string]string
	conditionalEdges map[string]ConditionalEdge[State]
	entryPoint       string
}

// Invoke executes the graph with the given initial state.
func (r *Runnable[State]) Invoke(ctx context.Context, initialState State) (State, error) {
	currentState := initialState
	currentNode := r.entryPoint

	// To prevent infinite loops in malformed graphs
	// We could add a MaxSteps configuration, but for simplicity we rely on the graph structure.
	// For production, a StepCounter check would be advised.

	for currentNode != END {
		// 1. Check if node exists
		nodeAction, exists := r.nodes[currentNode]
		if !exists {
			return currentState, fmt.Errorf("node '%s' not found in graph", currentNode)
		}

		// 2. Execute node action
		var err error
		currentState, err = nodeAction(ctx, currentState)
		if err != nil {
			return currentState, fmt.Errorf("error executing node '%s': %w", currentNode, err)
		}

		// 3. Determine next node
		nextNode := ""

		// Check conditional edges first
		if condEdge, hasCond := r.conditionalEdges[currentNode]; hasCond {
			conditionOutput, err := condEdge.Condition(ctx, currentState)
			if err != nil {
				return currentState, fmt.Errorf("error evaluating condition on node '%s': %w", currentNode, err)
			}
			
			if condEdge.Mapping != nil {
				mappedNode, mapped := condEdge.Mapping[conditionOutput]
				if !mapped {
					return currentState, fmt.Errorf("condition output '%s' not found in mapping on node '%s'", conditionOutput, currentNode)
				}
				nextNode = mappedNode
			} else {
				// If mapping is nil, use the output directly as the next node
				nextNode = conditionOutput
			}
		} else if normalEdge, hasEdge := r.edges[currentNode]; hasEdge {
			// Check normal edges
			nextNode = normalEdge
		} else {
			// If no outgoing edge is defined, we treat it as an error because
			// in LangGraph you explicitly route to END.
			return currentState, fmt.Errorf("node '%s' has no outgoing edges and is not END", currentNode)
		}

		currentNode = nextNode
	}

	return currentState, nil
}

// ToMermaid generates a Mermaid.js flowchart representation of the compiled graph.
func (r *Runnable[State]) ToMermaid() string {
	var sb strings.Builder
	sb.WriteString("graph TD;\n")
	
	// Add entry point
	sb.WriteString(fmt.Sprintf("    %s --> %s;\n", START, r.entryPoint))

	// Normal edges
	for from, to := range r.edges {
		sb.WriteString(fmt.Sprintf("    %s --> %s;\n", from, to))
	}

	// Conditional edges
	for from, condEdge := range r.conditionalEdges {
		if condEdge.Mapping != nil {
			for output, to := range condEdge.Mapping {
				// Escape outputs to avoid Mermaid syntax issues if they contain spaces
				sb.WriteString(fmt.Sprintf("    %s -. %s .-> %s;\n", from, output, to))
			}
		} else {
			sb.WriteString(fmt.Sprintf("    %s -. custom_logic .-> unknown;\n", from))
		}
	}

	return sb.String()
}

