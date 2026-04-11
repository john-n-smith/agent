package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/peterh/liner"
)

var promptInput = bufio.NewReader(os.Stdin)

func promptLine(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	line, err := promptInput.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptPath(prompt string) (string, error) {
	if !liner.TerminalSupported() {
		return promptLine(prompt)
	}

	state := liner.NewLiner()
	defer state.Close()
	state.SetCtrlCAborts(true)
	state.SetCompleter(completePathCandidates)

	line, err := state.Prompt(prompt)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptChoice(prompt string, allowed []string, defaultValue string) (string, error) {
	allowedSet := map[string]struct{}{}
	for _, value := range allowed {
		allowedSet[value] = struct{}{}
	}

	for {
		reply, err := promptLine(prompt)
		if err != nil {
			return "", err
		}
		reply = strings.ToLower(strings.TrimSpace(reply))
		if reply == "" {
			return defaultValue, nil
		}
		if _, ok := allowedSet[reply]; ok {
			return reply, nil
		}
		fmt.Fprintf(os.Stderr, "Enter %s.\n", strings.Join(allowed, " or "))
	}
}

func promptYesNo(prompt string) (bool, error) {
	reply, err := promptLine(prompt)
	if err != nil {
		return false, err
	}
	reply = strings.TrimSpace(reply)
	if reply == "" {
		return false, nil
	}
	return strings.EqualFold(reply, "y") || strings.EqualFold(reply, "yes"), nil
}

func completePathCandidates(input string) []string {
	rawDir, rawPrefix := filepath.Split(input)
	searchDir := rawDir
	if searchDir == "" {
		searchDir = "."
	}

	expandedSearchDir, err := expandHome(searchDir)
	if err != nil {
		return nil
	}

	entries, err := os.ReadDir(expandedSearchDir)
	if err != nil {
		return nil
	}

	var candidates []string
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, rawPrefix) {
			continue
		}

		candidate := rawDir + name
		if entry.IsDir() {
			candidate += string(os.PathSeparator)
		}
		candidates = append(candidates, candidate)
	}

	sort.Strings(candidates)
	return candidates
}
