package main

import (
	"bufio"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestPromptLineReadsSequentialLinesFromSharedInput(t *testing.T) {
	previousInput := promptInput
	promptInput = bufio.NewReader(strings.NewReader("first\nsecond\n"))
	t.Cleanup(func() {
		promptInput = previousInput
	})

	first, err := promptLine("first: ")
	if err != nil {
		t.Fatalf("promptLine() returned error: %v", err)
	}
	second, err := promptLine("second: ")
	if err != nil {
		t.Fatalf("promptLine() returned error: %v", err)
	}

	if first != "first" {
		t.Fatalf("first promptLine() = %q, want %q", first, "first")
	}
	if second != "second" {
		t.Fatalf("second promptLine() = %q, want %q", second, "second")
	}
}

func TestPromptPathWithDefaultReturnsDefaultOnEmptyInput(t *testing.T) {
	previousInput := promptInput
	promptInput = bufio.NewReader(strings.NewReader("\n"))
	t.Cleanup(func() {
		promptInput = previousInput
	})

	got, err := promptPathWithDefault("Workspace path: ", "/tmp/workspace")
	if err != nil {
		t.Fatalf("promptPathWithDefault() returned error: %v", err)
	}
	if got != "/tmp/workspace" {
		t.Fatalf("promptPathWithDefault() = %q, want %q", got, "/tmp/workspace")
	}
}

func TestPromptPathWithDefaultUsesEnteredValue(t *testing.T) {
	previousInput := promptInput
	promptInput = bufio.NewReader(strings.NewReader("/tmp/override\n"))
	t.Cleanup(func() {
		promptInput = previousInput
	})

	got, err := promptPathWithDefault("Workspace path: ", "/tmp/workspace")
	if err != nil {
		t.Fatalf("promptPathWithDefault() returned error: %v", err)
	}
	if got != "/tmp/override" {
		t.Fatalf("promptPathWithDefault() = %q, want %q", got, "/tmp/override")
	}
}

func TestDefaultWorkspacePromptPath(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	got := defaultWorkspacePromptPath()
	want := filepath.Join(homeDir, "git")
	if got != want {
		t.Fatalf("defaultWorkspacePromptPath() = %q, want %q", got, want)
	}
}

func TestDefaultHarnessPromptPath(t *testing.T) {
	got := defaultHarnessPromptPath("/tmp/workspace")
	want := filepath.Join("/tmp/workspace", "harbour-harness")
	if got != want {
		t.Fatalf("defaultHarnessPromptPath() = %q, want %q", got, want)
	}
}

func TestCompletePathCandidatesForAbsolutePath(t *testing.T) {
	base := t.TempDir()
	mustMkdirAll(t, filepath.Join(base, "alpha"))
	mustMkdirAll(t, filepath.Join(base, "alpine"))
	mustWriteFile(t, filepath.Join(base, "alphabet.txt"))
	mustWriteFile(t, filepath.Join(base, "beta.txt"))

	got := completePathCandidates(filepath.Join(base, "alp"))
	want := []string{
		filepath.Join(base, "alpha") + string(os.PathSeparator),
		filepath.Join(base, "alphabet.txt"),
		filepath.Join(base, "alpine") + string(os.PathSeparator),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("completePathCandidates() = %#v, want %#v", got, want)
	}
}

func TestCompletePathCandidatesForTildePath(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	mustMkdirAll(t, filepath.Join(homeDir, "git"))
	mustMkdirAll(t, filepath.Join(homeDir, "gist"))

	got := completePathCandidates("~/gi")
	want := []string{"~/gist/", "~/git/"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("completePathCandidates() = %#v, want %#v", got, want)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() returned error: %v", err)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("test\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
}
