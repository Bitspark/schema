package websocket

import (
	"crypto/tls"
	"testing"
	"time"
)

func TestWebSocketFactoryFunctions(t *testing.T) {
	t.Run("TLS configuration", func(t *testing.T) {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}

		builder := NewConfigBuilder().TLS(config)
		builtConfig := builder.Build()

		if builtConfig.TLSConfig != config {
			t.Error("Expected TLS config to be set")
		}
	})

	t.Run("SecureServerConfig", func(t *testing.T) {
		host := "0.0.0.0"
		port := 443
		tlsConfig := &tls.Config{InsecureSkipVerify: true}

		config := SecureServerConfig(host, port, tlsConfig)

		if config.TLSConfig != tlsConfig {
			t.Error("Expected TLS config to be set for secure server")
		}

		if config.Host != host {
			t.Errorf("Expected host '%s', got '%s'", host, config.Host)
		}

		if config.Port != port {
			t.Errorf("Expected port %d, got %d", port, config.Port)
		}
	})

	t.Run("DevelopmentConfig", func(t *testing.T) {
		config := DevelopmentConfig()

		if config.Host != "localhost" {
			t.Errorf("Expected host 'localhost', got '%s'", config.Host)
		}

		if config.Port != 8080 {
			t.Errorf("Expected port 8080, got %d", config.Port)
		}

		// Development config should have relaxed timeouts
		if config.ReadTimeout < 30*time.Second {
			t.Errorf("Expected longer read timeout for development, got %v", config.ReadTimeout)
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

		// Production config should be configured (just verify it's set)
		if config.ReadTimeout <= 0 {
			t.Error("Expected read timeout to be set for production")
		}

		if config.WriteTimeout <= 0 {
			t.Error("Expected write timeout to be set for production")
		}
	})

	t.Run("HighPerformanceConfig", func(t *testing.T) {
		host := "localhost"
		port := 8080
		config := HighPerformanceConfig(host, port)

		if config.Host != host {
			t.Errorf("Expected host '%s', got '%s'", host, config.Host)
		}

		if config.Port != port {
			t.Errorf("Expected port %d, got %d", port, config.Port)
		}

		if config.MaxMessageSize < 64*1024 {
			t.Errorf("Expected larger max message size for high performance, got %d", config.MaxMessageSize)
		}

		if config.ReadBufferSize < 4096 {
			t.Errorf("Expected larger read buffer for high performance, got %d", config.ReadBufferSize)
		}

		if config.WriteBufferSize < 4096 {
			t.Errorf("Expected larger write buffer for high performance, got %d", config.WriteBufferSize)
		}

		if !config.EnableCompression {
			t.Error("Expected compression to be enabled for high performance")
		}
	})

	t.Run("NewSecureServer", func(t *testing.T) {
		host := "localhost"
		port := 8443
		tlsConfig := &tls.Config{InsecureSkipVerify: true}

		portal := NewSecureServer(host, port, tlsConfig)

		if portal == nil {
			t.Fatal("Expected non-nil portal")
		}

		schemes := portal.Scheme()
		found := false
		for _, scheme := range schemes {
			if scheme == "wss" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected secure server to support 'wss' scheme")
		}
	})

	t.Run("NewDevelopmentPortal", func(t *testing.T) {
		portal := NewDevelopmentPortal()

		if portal == nil {
			t.Fatal("Expected non-nil portal")
		}

		// Should be ready to use
		schemes := portal.Scheme()
		if len(schemes) == 0 {
			t.Error("Expected development portal to have supported schemes")
		}
	})

	t.Run("NewProductionPortal", func(t *testing.T) {
		host := "0.0.0.0"
		port := 80
		portal := NewProductionPortal(host, port)

		if portal == nil {
			t.Fatal("Expected non-nil portal")
		}

		// Should be ready to use
		schemes := portal.Scheme()
		if len(schemes) == 0 {
			t.Error("Expected production portal to have supported schemes")
		}
	})

	t.Run("NewHighPerformancePortal", func(t *testing.T) {
		host := "localhost"
		port := 8080
		portal := NewHighPerformancePortal(host, port)

		if portal == nil {
			t.Fatal("Expected non-nil portal")
		}

		// Should be ready to use
		schemes := portal.Scheme()
		if len(schemes) == 0 {
			t.Error("Expected high performance portal to have supported schemes")
		}
	})

	t.Run("CreatePortalFromConfig", func(t *testing.T) {
		config := Config{
			Host: "localhost",
			Port: 9999,
		}

		portal := CreatePortalFromConfig(config)

		if portal == nil {
			t.Fatal("Expected non-nil portal")
		}

		// Should use the provided config
		schemes := portal.Scheme()
		if len(schemes) == 0 {
			t.Error("Expected portal to have supported schemes")
		}
	})

	t.Run("ClonePortal", func(t *testing.T) {
		original := NewDevelopmentPortal()
		cloned := ClonePortal(original)

		if cloned == nil {
			t.Fatal("Expected non-nil cloned portal")
		}

		if cloned == original {
			t.Error("Expected cloned portal to be different instance")
		}

		// Should have same schemes
		originalSchemes := original.Scheme()
		clonedSchemes := cloned.Scheme()

		if len(originalSchemes) != len(clonedSchemes) {
			t.Errorf("Expected same number of schemes, got %d vs %d", len(originalSchemes), len(clonedSchemes))
		}
	})

	t.Run("MergeConfigs", func(t *testing.T) {
		base := Config{
			Host:        "localhost",
			Port:        8080,
			ReadTimeout: 30 * time.Second,
		}

		override := Config{
			Port:         9999,
			WriteTimeout: 20 * time.Second,
		}

		merged := MergeConfigs(base, override)

		// Should keep base host
		if merged.Host != "localhost" {
			t.Errorf("Expected host 'localhost', got '%s'", merged.Host)
		}

		// Should use override port
		if merged.Port != 9999 {
			t.Errorf("Expected port 9999, got %d", merged.Port)
		}

		// Should keep base read timeout
		if merged.ReadTimeout != 30*time.Second {
			t.Errorf("Expected read timeout 30s, got %v", merged.ReadTimeout)
		}

		// Should use override write timeout
		if merged.WriteTimeout != 20*time.Second {
			t.Errorf("Expected write timeout 20s, got %v", merged.WriteTimeout)
		}
	})
}
