package graph

import (
	"context"
	"sync"
)

// ReducerFunc is a function that merges the results of parallel nodes.
// It receives the original state before the parallel execution, and a slice of new states
// returned by the parallel nodes.
type ReducerFunc[State any] func(ctx context.Context, originalState State, newStates []State) (State, error)

// Parallel creates a single NodeFunc that executes multiple NodeFuncs concurrently.
// It waits for all of them to finish, and then uses the reducer to merge their output states.
func Parallel[State any](reducer ReducerFunc[State], nodes ...NodeFunc[State]) NodeFunc[State] {
	return func(ctx context.Context, state State) (State, error) {
		if len(nodes) == 0 {
			return state, nil
		}

		var wg sync.WaitGroup
		
		results := make([]State, len(nodes))
		errors := make([]error, len(nodes))

		for i, node := range nodes {
			wg.Add(1)
			go func(index int, n NodeFunc[State]) {
				defer wg.Done()
				// Execute the node with the current state.
				// Note: If the state contains maps/slices, users should be careful 
				// not to mutate them directly inside the parallel nodes to avoid race conditions,
				// or they should use mutexes within the state.
				res, err := n(ctx, state)
				results[index] = res
				errors[index] = err
			}(i, node)
		}

		wg.Wait()

		// Check if any of the parallel nodes returned an error
		for _, err := range errors {
			if err != nil {
				return state, err // Return the first error encountered
			}
		}

		// Merge the results using the provided reducer
		return reducer(ctx, state, results)
	}
}
