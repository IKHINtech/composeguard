package state

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/IKHINtech/composeguard/internal/checker"
)

type NotificationDecision struct {
	ShouldSend bool
	Reason     string
	Problems   []checker.Result
}

func EvaluateNotification(
	currentState *State,
	results []checker.Result,
	cooldownMinutes int,
	now time.Time,
) NotificationDecision {
	if currentState == nil {
		currentState = &State{
			Problems: make(map[string]ProblemState),
		}
	}
	if currentState.Problems == nil {
		currentState.Problems = make(map[string]ProblemState)
	}

	if cooldownMinutes <= 0 {
		cooldownMinutes = 60
	}

	problems := problemResults(results)
	if len(problems) == 0 {
		return NotificationDecision{
			ShouldSend: false,
			Reason:     "no problem found",
			Problems:   nil,
		}
	}

	cooldown := time.Duration(cooldownMinutes) * time.Minute

	shouldSend := false
	reasons := make([]string, 0)

	for _, problem := range problems {
		key := problemKey(problem)
		fingerprint := problemFingerprint(problem)

		previous, exists := currentState.Problems[key]
		if !exists {
			shouldSend = true
			reasons = append(reasons, fmt.Sprintf("%s is a new problem", problem.Name))
			continue
		}

		if previous.Fingerprint != fingerprint {
			shouldSend = true
			reasons = append(reasons, fmt.Sprintf("%s changed", problem.Name))
			continue
		}

		if previous.LastSentAt.IsZero() || now.Sub(previous.LastSentAt) >= cooldown {
			shouldSend = true
			reasons = append(reasons, fmt.Sprintf("%s cooldown expired", problem.Name))
			continue
		}
	}

	if !shouldSend {
		return NotificationDecision{
			ShouldSend: false,
			Reason:     fmt.Sprintf("same problem still in cooldown (%d minutes)", cooldownMinutes),
			Problems:   problems,
		}
	}

	return NotificationDecision{
		ShouldSend: true,
		Reason:     strings.Join(reasons, "; "),
		Problems:   problems,
	}
}

func UpdateAfterNotification(
	currentState *State,
	problems []checker.Result,
	now time.Time,
	notificationSent bool,
) *State {
	if currentState == nil {
		currentState = &State{
			Problems: make(map[string]ProblemState),
		}
	}

	if currentState.Problems == nil {
		currentState.Problems = make(map[string]ProblemState)
	}

	currentProblemKeys := make(map[string]struct{}, len(problems))

	for _, problem := range problems {
		key := problemKey(problem)
		fingerprint := problemFingerprint(problem)

		currentProblemKeys[key] = struct{}{}

		previous := currentState.Problems[key]

		if notificationSent {
			previous.LastSentAt = now
		}

		previous.Fingerprint = fingerprint
		previous.Status = string(problem.Status)
		previous.Message = problem.Message
		previous.LastSeenAt = now

		currentState.Problems[key] = previous
	}

	// v0.5.0 behavior:
	// If a problem no longer exists, remove it from state.
	// v0.6.0 will use this transition for recovery notification.
	for key := range currentState.Problems {
		if _, exists := currentProblemKeys[key]; !exists {
			delete(currentState.Problems, key)
		}
	}

	return currentState
}

func problemResults(results []checker.Result) []checker.Result {
	problems := make([]checker.Result, 0)

	for _, result := range results {
		if result.Name == "" && result.Message == "" && result.Status == "" {
			continue
		}

		if result.Status == checker.StatusCritical || result.Status == checker.StatusWarning {
			problems = append(problems, result)
		}
	}

	return problems
}

func problemKey(result checker.Result) string {
	return strings.TrimSpace(result.Name)
}

func problemFingerprint(result checker.Result) string {
	source := fmt.Sprintf("%s|%s|%s", result.Name, result.Status, result.Message)
	sum := sha256.Sum256([]byte(source))
	return hex.EncodeToString(sum[:])
}
