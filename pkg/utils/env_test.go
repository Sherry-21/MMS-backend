package utils

import (
	"os"
	"testing"
	"time"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "custom_value")
	defer os.Unsetenv("TEST_KEY")

	if val := GetEnv("TEST_KEY", "default"); val != "custom_value" {
		t.Errorf("expected 'custom_value', got '%s'", val)
	}

	if val := GetEnv("NON_EXISTENT_KEY", "default"); val != "default" {
		t.Errorf("expected 'default', got '%s'", val)
	}
}

func TestGetEnvAsInt(t *testing.T) {
	os.Setenv("TEST_INT", "9090")
	defer os.Unsetenv("TEST_INT")

	if val := GetEnvAsInt("TEST_INT", 8080); val != 9090 {
		t.Errorf("expected 9090, got %d", val)
	}

	if val := GetEnvAsInt("NON_EXISTENT", 8080); val != 8080 {
		t.Errorf("expected 8080, got %d", val)
	}
}

func TestGetEnvAsBool(t *testing.T) {
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	if val := GetEnvAsBool("TEST_BOOL", false); !val {
		t.Errorf("expected true, got false")
	}

	if val := GetEnvAsBool("NON_EXISTENT", true); !val {
		t.Errorf("expected true, got false")
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	os.Setenv("TEST_DUR", "30s")
	defer os.Unsetenv("TEST_DUR")

	if val := GetEnvAsDuration("TEST_DUR", 10*time.Second); val != 30*time.Second {
		t.Errorf("expected 30s, got %v", val)
	}
}

func TestGetEnvAsSlice(t *testing.T) {
	os.Setenv("TEST_SLICE", "a, b, c")
	defer os.Unsetenv("TEST_SLICE")

	res := GetEnvAsSlice("TEST_SLICE", []string{"default"}, ",")
	if len(res) != 3 || res[0] != "a" || res[1] != "b" || res[2] != "c" {
		t.Errorf("unexpected slice result: %v", res)
	}
}
