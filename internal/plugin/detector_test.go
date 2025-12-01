package plugin

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGoDetector_Detect(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) string
		wantDetected  bool
		wantVersion   string
		wantModule    string
		wantWorkspace bool
	}{
		{
			name: "detects go workspace",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				goModContent := `module github.com/example/test

go 1.24
`
				goWorkContent := `go 1.24

use (
	./module1
	./module2
)
`
				err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)
				if err != nil {
					t.Fatal(err)
				}
				err = os.WriteFile(filepath.Join(tmpDir, "go.work"), []byte(goWorkContent), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			wantDetected:  true,
			wantVersion:   "1.24",
			wantModule:    "github.com/example/test",
			wantWorkspace: true,
		},
		{
			name: "does not detect non-go project",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create a package.json instead
				err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"name": "test"}`), 0644)
				if err != nil {
					t.Fatal(err)
				}
				return tmpDir
			},
			wantDetected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewGoDetector()
			projectPath := tt.setupFunc(t)

			result, err := detector.Detect(context.Background(), projectPath)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if !tt.wantDetected {
				if result != nil {
					t.Errorf("Detect() should return nil for non-Go project, got %v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Detect() returned nil for Go project")
			}

			data, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("Detect() result is not map[string]interface{}, got %T", result)
			}

			if detected, ok := data["detected"].(bool); !ok || !detected {
				t.Errorf("Detect() detected = %v, want true", detected)
			}

			if tt.wantVersion != "" {
				if version, ok := data["go_version"].(string); !ok || version != tt.wantVersion {
					t.Errorf("Detect() go_version = %v, want %v", version, tt.wantVersion)
				}
			}

			if tt.wantModule != "" {
				if module, ok := data["module"].(string); !ok || module != tt.wantModule {
					t.Errorf("Detect() module = %v, want %v", module, tt.wantModule)
				}
			}

			if tt.wantWorkspace {
				if workspace, ok := data["workspace"].(bool); !ok || !workspace {
					t.Errorf("Detect() workspace = %v, want true", workspace)
				}
			}
		})
	}
}

func TestNewGoDetector(t *testing.T) {
	detector := NewGoDetector()

	if detector == nil {
		t.Fatal("NewGoDetector() returned nil")
	}

	if detector.base == nil {
		t.Fatal("NewGoDetector() base detector is nil")
	}
}

func TestGoDetector_Name(t *testing.T) {
	detector := NewGoDetector()

	if name := detector.Name(); name != "go" {
		t.Errorf("Name() = %v, want go", name)
	}
}

func TestGoDetector_Merge(t *testing.T) {
	detector := NewGoDetector()

	existing := map[string]interface{}{"old": "data"}
	new := map[string]interface{}{"new": "data"}

	// Test that new data is preferred
	result, err := detector.Merge(existing, new)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if !reflect.DeepEqual(result, new) {
		t.Errorf("Merge() should prefer new data, got %v, want %v", result, new)
	}

	// Test that existing is returned when new is nil
	result, err = detector.Merge(existing, nil)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	if !reflect.DeepEqual(result, existing) {
		t.Errorf("Merge() should return existing when new is nil, got %v, want %v", result, existing)
	}
}
