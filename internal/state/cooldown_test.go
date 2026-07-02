package state

import (
	"testing"
	"time"

	"github.com/IKHINtech/composeguard/internal/checker"
)

func TestEvaluateNotificationNewProblem(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	state := &State{
		Problems: map[string]ProblemState{},
	}

	results := []checker.Result{
		{
			Name:    "Disk: /",
			Status:  checker.StatusCritical,
			Message: "98% used",
		},
	}

	decision := EvaluateNotification(state, results, 60, now)

	if !decision.ShouldSend {
		t.Fatalf("expected should send, got false: %s", decision.Reason)
	}
}

func TestEvaluateNotificationSameProblemInCooldown(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	result := checker.Result{
		Name:    "Disk: /",
		Status:  checker.StatusCritical,
		Message: "98% used",
	}

	state := &State{
		Problems: map[string]ProblemState{
			"Disk: /": {
				Fingerprint: problemFingerprint(result),
				Status:      string(result.Status),
				Message:     result.Message,
				LastSentAt:  now.Add(-30 * time.Minute),
				LastSeenAt:  now.Add(-30 * time.Minute),
			},
		},
	}

	decision := EvaluateNotification(state, []checker.Result{result}, 60, now)

	if decision.ShouldSend {
		t.Fatalf("expected should not send, got true: %s", decision.Reason)
	}
}

func TestEvaluateNotificationSameProblemAfterCooldown(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	result := checker.Result{
		Name:    "Disk: /",
		Status:  checker.StatusCritical,
		Message: "98% used",
	}

	state := &State{
		Problems: map[string]ProblemState{
			"Disk: /": {
				Fingerprint: problemFingerprint(result),
				Status:      string(result.Status),
				Message:     result.Message,
				LastSentAt:  now.Add(-2 * time.Hour),
				LastSeenAt:  now.Add(-2 * time.Hour),
			},
		},
	}

	decision := EvaluateNotification(state, []checker.Result{result}, 60, now)

	if !decision.ShouldSend {
		t.Fatalf("expected should send after cooldown, got false: %s", decision.Reason)
	}
}

func TestEvaluateNotificationProblemChanged(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	oldResult := checker.Result{
		Name:    "Disk: /",
		Status:  checker.StatusWarning,
		Message: "85% used",
	}

	newResult := checker.Result{
		Name:    "Disk: /",
		Status:  checker.StatusCritical,
		Message: "98% used",
	}

	state := &State{
		Problems: map[string]ProblemState{
			"Disk: /": {
				Fingerprint: problemFingerprint(oldResult),
				Status:      string(oldResult.Status),
				Message:     oldResult.Message,
				LastSentAt:  now.Add(-10 * time.Minute),
				LastSeenAt:  now.Add(-10 * time.Minute),
			},
		},
	}

	decision := EvaluateNotification(state, []checker.Result{newResult}, 60, now)

	if !decision.ShouldSend {
		t.Fatalf("expected should send because problem changed, got false: %s", decision.Reason)
	}
}

func TestUpdateAfterNotificationRemovesResolvedProblems(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	state := &State{
		Problems: map[string]ProblemState{
			"Disk: /": {
				Fingerprint: "abc",
				Status:      "CRITICAL",
				Message:     "98% used",
				LastSentAt:  now.Add(-1 * time.Hour),
				LastSeenAt:  now.Add(-1 * time.Hour),
			},
		},
	}

	updated := UpdateAfterNotification(state, nil, now, false)

	if len(updated.Problems) != 0 {
		t.Fatalf("expected resolved problem to be removed")
	}
}
