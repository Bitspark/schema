package tests

import (
	"context"
	"defs.dev/schema/construct/builders"
	"defs.dev/schema/runtime/portal"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"defs.dev/schema/api"
	"defs.dev/schema/core"
)

// Simple WebSocket test to debug connection issues
func TestE2E_WebSocketSimple(t *testing.T) {
	ctx := context.Background()

	// Create WebSocket portal with minimal config
	wsPortal := portal.NewWebSocketPortal(&portal.WebSocketConfig{
		Host:           "localhost",
		Port:           8090,
		Path:           "/ws",
		PingPeriod:     54 * time.Second,
		PongWait:       60 * time.Second,
		WriteTimeout:   10 * time.Second,
		ReadTimeout:    60 * time.Second,
		MaxMessageSize: 512 * 1024,
	}, nil, nil) // Use default registries

	// Register a simple function
	simpleFunc := &SimpleTestFunction{
		name: "echo",
		schema: builders.NewFunctionSchema().
			Name("echo").
			Input("message", builders.NewStringSchema().Build()).
			RequiredInputs("message").
			Output("echo", builders.NewStringSchema().Build()).
			Build(),
		handler: func(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
			message, _ := params.Get("message")
			return api.NewFunctionData(map[string]any{
				"echo": fmt.Sprintf("Echo: %s", message.(string)),
			}), nil
		},
	}

	_, err := wsPortal.Apply(ctx, simpleFunc)
	if err != nil {
		t.Fatalf("Failed to register function: %v", err)
	}

	// Start WebSocket server
	err = wsPortal.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start WebSocket portal: %v", err)
	}
	defer wsPortal.Stop(ctx)

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// Test connection
	u := url.URL{Scheme: "ws", Host: "localhost:8090", Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Send a simple message
	message := map[string]any{
		"type":     "call",
		"id":       "test1",
		"function": "echo",
		"data": map[string]any{
			"message": "Hello WebSocket!",
		},
	}

	err = conn.WriteJSON(message)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Read response with timeout
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var response map[string]any
	err = conn.ReadJSON(&response)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	t.Logf("Received response: %+v", response)

	// Validate response
	if response["type"] != "response" {
		t.Errorf("Expected response type 'response', got %v", response["type"])
	}

	if response["id"] != "test1" {
		t.Errorf("Expected response id 'test1', got %v", response["id"])
	}

	if data, ok := response["data"].(map[string]any); ok {
		if result, ok := data["result"].(map[string]any); ok {
			if echo, ok := result["echo"].(string); ok {
				if echo != "Echo: Hello WebSocket!" {
					t.Errorf("Expected 'Echo: Hello WebSocket!', got %s", echo)
				}
			} else {
				t.Error("Expected echo field in result")
			}
		} else {
			t.Error("Expected result field in data")
		}
	} else {
		t.Error("Expected data field in response")
	}

	t.Logf("Simple WebSocket test passed!")
}

type SimpleTestFunction struct {
	name    string
	schema  core.FunctionSchema
	handler func(ctx context.Context, params api.FunctionData) (api.FunctionData, error)
}

func (f *SimpleTestFunction) Name() string {
	return f.name
}

func (f *SimpleTestFunction) Schema() core.FunctionSchema {
	return f.schema
}

func (f *SimpleTestFunction) Call(ctx context.Context, params api.FunctionData) (api.FunctionData, error) {
	return f.handler(ctx, params)
}
