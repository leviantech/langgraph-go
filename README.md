# LangGraph Go

A lightweight, strongly-typed implementation of the core features of [LangGraph](https://github.com/langchain-ai/langgraph) in Golang, utilizing Go 1.18+ Generics for type-safe state management.

## Features

- **Generic State**: Define your own state struct and pass it safely between nodes without type assertions.
- **Nodes & Edges**: Build workflows as graphs of functions.
- **Conditional Routing**: Use custom logic to branch the execution path.
- **Synchronous Execution**: Simple and predictable `Invoke` method for running the graph.

## Installation

```bash
go get github.com/leviantech/langgraph-go
```

## Quick Start

See the `examples/` directory for full, runnable examples.

### Basic Workflow

```go
package main

import (
	"context"
	"fmt"
	"github.com/leviantech/langgraph-go/graph"
)

type State struct {
	Value string
}

func NodeA(ctx context.Context, state State) (State, error) {
	state.Value += "A"
	return state, nil
}

func NodeB(ctx context.Context, state State) (State, error) {
	state.Value += "B"
	return state, nil
}

func main() {
	g := graph.NewStateGraph[State]()

	g.AddNode("A", NodeA)
	g.AddNode("B", NodeB)

	g.AddEdge("A", "B")
	g.SetEntryPoint("A")
	g.SetFinishPoint("B")

	app, _ := g.Compile()

	finalState, _ := app.Invoke(context.Background(), State{Value: ""})
	fmt.Println(finalState.Value) // Output: AB
}
```

### Conditional Routing

You can route execution using `AddConditionalEdge`:

```go
g.AddConditionalEdge("StartNode", func(ctx context.Context, state State) (string, error) {
    if state.Value == "win" {
        return "win_node", nil
    }
    return "lose_node", nil
}, map[string]string{
    "win_node":  "WinNode",
    "lose_node": "LoseNode",
})
```

## Architecture

This project is a clean-room implementation inspired by the Python `langgraph` package, adapted for Go's idioms. 
The use of Generics `[State any]` ensures you don't have to deal with `map[string]any` if you don't want to, allowing for strict type safety throughout your workflow execution.
