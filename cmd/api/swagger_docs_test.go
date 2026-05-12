package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSwaggerDocsUpToDate(t *testing.T) {
	swagPath, err := exec.LookPath("swag")
	if err != nil {
		t.Fatal("swag CLI not installed. install with: go install github.com/swaggo/swag/cmd/swag@latest")
	}

	tmpDir := t.TempDir()
	cmd := exec.Command(swagPath,
		"init",
		"--parseDependency",
		"--parseInternal",
		"-g",
		"docs.go",
		"-d",
		".,../../api",
		"-o",
		tmpDir,
		"--packageName",
		"docs",
	)
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to generate swagger docs: %v\n%s", err, string(output))
	}

	compareFile(t, filepath.Join("..", "..", "docs", "swagger.json"), filepath.Join(tmpDir, "swagger.json"))
	compareFile(t, filepath.Join("..", "..", "docs", "swagger.yaml"), filepath.Join(tmpDir, "swagger.yaml"))
	compareFile(t, filepath.Join("..", "..", "docs", "docs.go"), filepath.Join(tmpDir, "docs.go"))
}

func compareFile(t *testing.T, committedPath, generatedPath string) {
	t.Helper()

	committed, err := os.ReadFile(committedPath)
	if err != nil {
		t.Fatalf("failed to read committed file %s: %v", committedPath, err)
	}
	generated, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatalf("failed to read generated file %s: %v", generatedPath, err)
	}
	if !bytes.Equal(committed, generated) {
		t.Fatalf("%s is out of date; run go generate ./cmd/api and commit updated docs", committedPath)
	}
}
