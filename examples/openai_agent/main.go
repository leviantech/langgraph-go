package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/leviantech/langgraph-go/graph"
	"github.com/sashabaranov/go-openai"
)

// AgentState maintains the chat history.
type AgentState struct {
	Messages []openai.ChatCompletionMessage
}

// AgentNode is the node responsible for calling the OpenAI LLM.
func AgentNode(ctx context.Context, state AgentState) (AgentState, error) {
	fmt.Println("🤖 Agent is thinking...")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "dummy_key" // Some compatible APIs require a non-empty string
	}

	config := openai.DefaultConfig(apiKey)
	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}
	client := openai.NewClientWithConfig(config)

	// Define our tools (Function Calling)
	tools := []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get the current weather in a given location",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"location": {
							"type": "string",
							"description": "The city and state, e.g. Tokyo, JP"
						}
					},
					"required": ["location"]
				}`),
			},
		},
	}

	modelName := os.Getenv("OPENAI_MODEL")
	if modelName == "" {
		modelName = openai.GPT4oMini // Default to GPT-4o-mini
	}

	// Call OpenAI API
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    modelName,
			Messages: state.Messages,
			Tools:    tools,
		},
	)

	if err != nil {
		return state, err
	}

	// Append the assistant's response to the state
	assistantMsg := resp.Choices[0].Message
	state.Messages = append(state.Messages, assistantMsg)

	return state, nil
}

// ToolNode is the node responsible for executing tools requested by the LLM.
func ToolNode(ctx context.Context, state AgentState) (AgentState, error) {
	fmt.Println("🛠️  Executing tool...")
	lastMessage := state.Messages[len(state.Messages)-1]

	if len(lastMessage.ToolCalls) == 0 {
		return state, nil
	}

	for _, toolCall := range lastMessage.ToolCalls {
		if toolCall.Function.Name == "get_weather" {
			// In a real app, you would parse the arguments and call a real weather API
			var args map[string]interface{}
			_ = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
			location := args["location"]

			fmt.Printf("   -> Looking up weather for: %s\n", location)

			// Mocking the result
			weatherResult := fmt.Sprintf("The weather in %s is sunny and 25°C.", location)

			// Create the Tool response message
			toolMessage := openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    weatherResult,
				Name:       toolCall.Function.Name,
				ToolCallID: toolCall.ID,
			}
			state.Messages = append(state.Messages, toolMessage)
		}
	}

	return state, nil
}

// ShouldContinue evaluates the state and decides the next step.
// If the LLM made a tool call, we go to "tools". Otherwise, we are done ("end").
func ShouldContinue(ctx context.Context, state AgentState) (string, error) {
	lastMessage := state.Messages[len(state.Messages)-1]

	if len(lastMessage.ToolCalls) > 0 {
		return "tools", nil
	}

	return "end", nil
}

func main() {
	if os.Getenv("OPENAI_API_KEY") == "" && os.Getenv("OPENAI_BASE_URL") == "" {
		log.Println("Skipping example: OPENAI_API_KEY or OPENAI_BASE_URL environment variable is not set")
		log.Println("Run with: OPENAI_API_KEY=your-key go run main.go")
		log.Println("Or for compatible APIs: OPENAI_BASE_URL=http://localhost:11434/v1 OPENAI_MODEL=llama3 go run main.go")
		return
	}

	// 1. Initialize the graph
	g := graph.NewStateGraph[AgentState]()

	// 2. Add nodes
	g.AddNode("agent", AgentNode)
	g.AddNode("tools", ToolNode)

	// 3. Define Entry Point
	g.SetEntryPoint("agent")

	// 4. Add Conditional Routing
	g.AddConditionalEdge("agent", ShouldContinue, map[string]string{
		"tools": "tools",
		"end":   graph.END, // If we map to graph.END, it tells the runner to stop
	})

	// 5. Add Normal Edge (loop back to agent after tools finish)
	g.AddEdge("tools", "agent")

	// 6. Compile
	app, err := g.Compile()
	if err != nil {
		log.Fatalf("Failed to compile graph: %v", err)
	}

	// 7. Execute Graph
	initialState := AgentState{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Hi, what is the weather in Tokyo right now?",
			},
		},
	}

	fmt.Println("User: Hi, what is the weather in Tokyo right now?")

	finalState, err := app.Invoke(context.Background(), initialState)
	if err != nil {
		log.Fatalf("Error invoking graph: %v", err)
	}

	// Print the final answer
	lastMsg := finalState.Messages[len(finalState.Messages)-1]
	fmt.Printf("\n🤖 Final Answer: %s\n", lastMsg.Content)
}
