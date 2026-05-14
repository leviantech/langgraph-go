package main

import (
	"context"
	"fmt"
	"log"

	"github.com/leviantech/langgraph-go/graph"
)

// 1. Definisikan State sederhana
type State struct {
	Results []string
}

// 2. Node A (Berjalan Paralel)
func NodeA(ctx context.Context, state State) (State, error) {
	fmt.Println("-> Menjalankan Node A")
	return State{Results: []string{"Data dari A"}}, nil
}

// 3. Node B (Berjalan Paralel)
func NodeB(ctx context.Context, state State) (State, error) {
	fmt.Println("-> Menjalankan Node B")
	return State{Results: []string{"Data dari B"}}, nil
}

// 4. Reducer untuk menggabungkan hasil dari Node A dan Node B
func MergeResults(ctx context.Context, originalState State, newStates []State) (State, error) {
	mergedState := originalState
	for _, s := range newStates {
		mergedState.Results = append(mergedState.Results, s.Results...)
	}
	return mergedState, nil
}

func main() {
	g := graph.NewStateGraph[State]()

	// Bungkus NodeA dan NodeB menjadi satu eksekusi paralel
	parallelNode := graph.Parallel(MergeResults, NodeA, NodeB)

	// Daftarkan ke graph
	g.AddNode("TaskParalel", parallelNode)
	
	// Set alur graph (karena hanya 1 step, langsung dari Entry ke TaskParalel lalu END)
	g.SetEntryPoint("TaskParalel")
	g.SetFinishPoint("TaskParalel")

	app, err := g.Compile()
	if err != nil {
		log.Fatal(err)
	}

	// (Opsional) Cetak visualisasi graph dalam format Mermaid
	fmt.Println("=== Visualisasi Graph ===")
	fmt.Println(app.ToMermaid())

	// Eksekusi Graph
	fmt.Println("=== Eksekusi Graph ===")
	initialState := State{Results: []string{"Data Awal"}}
	finalState, _ := app.Invoke(context.Background(), initialState)

	fmt.Printf("\nHasil Akhir: %v\n", finalState.Results)
}
