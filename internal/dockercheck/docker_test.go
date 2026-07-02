package dockercheck

import (
	"testing"

	"github.com/IKHINtech/composeguard/internal/checker"
)

func TestCheckContainersFindsMatchingContainerBeyondFirstEntry(t *testing.T) {
	original := listContainersFn
	t.Cleanup(func() {
		listContainersFn = original
	})

	listContainersFn = func() ([]dockerContainer, error) {
		return []dockerContainer{
			{Names: "unrelated", Status: "Up 1 minute", State: "running"},
			{Names: "mysql-container", Status: "Up 6 minutes", State: "running"},
			{Names: "my-postgres-db", Status: "Up 6 minutes", State: "running"},
		}, nil
	}

	results := CheckContainers([]string{"mysql-container", "my-postgres-db"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Status != checker.StatusOK {
			t.Fatalf("expected status OK, got %s for %s with message %q", result.Status, result.Name, result.Message)
		}
	}
}

func TestCheckContainersReportsMissingContainer(t *testing.T) {
	original := listContainersFn
	t.Cleanup(func() {
		listContainersFn = original
	})

	listContainersFn = func() ([]dockerContainer, error) {
		return []dockerContainer{
			{Names: "mysql-container", Status: "Up 6 minutes", State: "running"},
		}, nil
	}

	results := CheckContainers([]string{"my-postgres-db"})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Status != checker.StatusCritical {
		t.Fatalf("expected critical status, got %s", results[0].Status)
	}

	if results[0].Message != "container not found" {
		t.Fatalf("unexpected message %q", results[0].Message)
	}
}
