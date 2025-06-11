package http

import (
	"testing"
	"time"
)

func TestHTTPFactoryFunctions(t *testing.T) {
	t.Run("DevelopmentConfig", func(t *testing.T) {
		config := DevelopmentConfig()

		if config.Host != "localhost" {
			t.Errorf("Expected host 'localhost', got '%s'", config.Host)
		}

		if config.Port != 8080 {
			t.Errorf("Expected port 8080, got %d", config.Port)
		}

		// Development config should have timeouts configured
		if config.DefaultTimeout <= 0 {
			t.Error("Expected timeout to be configured for development")
		}
	})

	t.Run("ProductionConfig", func(t *testing.T) {
		host := "0.0.0.0"
		port := 80
		config := ProductionConfig(host, port)

		if config.Host != host {
			t.Errorf("Expected host '%s', got '%s'", host, config.Host)
		}

		if config.Port != port {
			t.Errorf("Expected port %d, got %d", port, config.Port)
		}

		// Production config should have timeouts configured
		if config.DefaultTimeout <= 0 {
			t.Error("Expected timeout to be configured for production")
		}
	})

	t.Run("RetryDelay", func(t *testing.T) {
		delay := 500 * time.Millisecond

		config := NewConfigBuilder().
			RetryDelay(delay).
			Build()

		if config.RetryDelay != delay {
			t.Errorf("Expected retry delay %v, got %v", delay, config.RetryDelay)
		}
	})

	t.Run("WithConfig", func(t *testing.T) {
		builder := WithConfig()

		if builder == nil {
			t.Fatal("Expected non-nil config builder")
		}

		config := builder.Host("example.com").Port(9999).Build()

		if config.Host != "example.com" {
			t.Errorf("Expected host 'example.com', got '%s'", config.Host)
		}

		if config.Port != 9999 {
			t.Errorf("Expected port 9999, got %d", config.Port)
		}
	})
}
