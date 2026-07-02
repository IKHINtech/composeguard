// Package state...
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type State struct {
	Problems map[string]ProblemState `json:"problems"`
}

type ProblemState struct {
	Fingerprint string    `json:"fingerprint"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	LastSentAt  time.Time `json:"last_sent_at"`
	LastSeenAt  time.Time `json:"last_seen_at"`
}

func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil && home == "" {
		return ".composeguard/state.json"
	}

	return filepath.Join(home, ".composeguard", "state.json")
}

func Load(path string) (*State, error) {
	if path == "" {
		path = DefaultPath()
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{
				Problems: make(map[string]ProblemState),
			}, nil
		}
		return nil, fmt.Errorf("failed to read state file %w", err)
	}

	var state State
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	if state.Problems == nil {
		state.Problems = make(map[string]ProblemState)
	}

	return &state, nil
}

func Save(path string, state *State) error {
	if path == "" {
		path = DefaultPath()
	}

	if state == nil {
		state = &State{
			Problems: make(map[string]ProblemState),
		}
	}
	if state.Problems == nil {
		state.Problems = make(map[string]ProblemState)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	raw, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		fmt.Printf("failed to marshal state: %v\n", err)
	}

	if err := os.WriteFile(path, raw, 0644); err != nil {
		return fmt.Errorf("failed to write state file %s: %w", path, err)
	}

	return nil
}
