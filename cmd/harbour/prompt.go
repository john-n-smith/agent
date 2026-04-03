package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func promptLine(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && len(line) == 0 {
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
