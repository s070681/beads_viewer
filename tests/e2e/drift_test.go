package main_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEndToEndDriftWorkflow(t *testing.T) {
	// 1. Build the binary
	tempDir := t.TempDir()
	binPath := filepath.Join(tempDir, "bv")

	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/bv/main.go")
	cmd.Dir = "../../"
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Build failed: %v\n%s", err, out)
	}

	// 2. Prepare environment
	envDir := filepath.Join(tempDir, "env")
	if err := os.MkdirAll(filepath.Join(envDir, ".beads"), 0755); err != nil {
		t.Fatal(err)
	}

	// 3. Create initial healthy graph
	// A -> B -> C
	jsonlContent := `{"id": "A", "title": "Task A", "status": "open", "priority": 1, "issue_type": "task"}
{"id": "B", "title": "Task B", "status": "open", "priority": 1, "issue_type": "task", "dependencies": [{"depends_on_id": "A", "type": "blocks"}]}
{"id": "C", "title": "Task C", "status": "open", "priority": 1, "issue_type": "task", "dependencies": [{"depends_on_id": "B", "type": "blocks"}]}`

	beadsPath := filepath.Join(envDir, ".beads", "beads.jsonl")
	if err := os.WriteFile(beadsPath, []byte(jsonlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 4. Save Baseline
	cmdSave := exec.Command(binPath, "--save-baseline", "Initial state")
	cmdSave.Dir = envDir
	if out, err := cmdSave.CombinedOutput(); err != nil {
		t.Fatalf("Save baseline failed: %v\n%s", err, out)
	}

	// Verify baseline file exists
	baselinePath := filepath.Join(envDir, ".bv", "baseline.json")
	if _, err := os.Stat(baselinePath); os.IsNotExist(err) {
		t.Fatalf("Baseline file not created at %s", baselinePath)
	}

	// 5. Check Drift (Should be clean)
	cmdCheck := exec.Command(binPath, "--check-drift")
	cmdCheck.Dir = envDir
	if out, err := cmdCheck.CombinedOutput(); err != nil {
		t.Fatalf("Check drift (clean) failed: %v\n%s", err, out)
	}

	// 6. Introduce Drift (New Cycle)
	// Create cycle A -> C -> B -> A
	driftedContent := `{"id": "A", "title": "Task A", "status": "open", "priority": 1, "issue_type": "task", "dependencies": [{"depends_on_id": "C", "type": "blocks"}]}
{"id": "B", "title": "Task B", "status": "open", "priority": 1, "issue_type": "task", "dependencies": [{"depends_on_id": "A", "type": "blocks"}]}
{"id": "C", "title": "Task C", "status": "open", "priority": 1, "issue_type": "task", "dependencies": [{"depends_on_id": "B", "type": "blocks"}]}`

	if err := os.WriteFile(beadsPath, []byte(driftedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 7. Check Drift (Should Fail with Exit Code 1)
	cmdDrift := exec.Command(binPath, "--check-drift")
	cmdDrift.Dir = envDir
	out, err := cmdDrift.CombinedOutput()

	// Expect exit code 1 (Critical)
	if err == nil {
		t.Fatalf("Expected drift check to fail, but it succeeded. Output:\n%s", out)
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got %T: %v", err, err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("Expected exit code 1 (critical), got %d", exitErr.ExitCode())
	}
	// Verify output contains "CRITICAL"
	outputStr := string(out)
	if !strings.Contains(outputStr, "CRITICAL") {
		t.Errorf("Expected output to mention CRITICAL, got:\n%s", outputStr)
	}

	// 8. Verify with JSON output
	cmdJson := exec.Command(binPath, "--check-drift", "--robot-drift")
	cmdJson.Dir = envDir
	outJson, err := cmdJson.CombinedOutput()
	// JSON mode also exits with code, so we expect error
	if err == nil {
		t.Fatal("Expected JSON drift check to fail with exit code")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(outJson, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\n%s", err, outJson)
	}

	if hasDrift, ok := result["has_drift"].(bool); !ok || !hasDrift {
		t.Error("JSON output has_drift should be true")
	}

	alerts, ok := result["alerts"].([]interface{})
	if !ok || len(alerts) == 0 {
		t.Error("JSON output should have alerts")
	} else {
		// Check first alert is new_cycle
		firstAlert := alerts[0].(map[string]interface{})
		if firstAlert["type"] != "new_cycle" {
			t.Errorf("Expected new_cycle alert, got %v", firstAlert["type"])
		}
	}
}
