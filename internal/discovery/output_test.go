package discovery

import "testing"

func TestFormatContainersYAML(t *testing.T) {
	containers := []DockerContainer{
		{
			Name:   "nginx",
			State:  "running",
			Status: "Up 1 hour",
		},
		{
			Name:   "postgres",
			State:  "running",
			Status: "Up 1 hour",
		},
	}

	expected := `docker:
  containers:
    - nginx
    - postgres
`

	actual := FormatContainersYAML(containers)
	if actual != expected {
		t.Fatalf("expected:\n%s\nactual:\n%s", expected, actual)
	}
}

func TestFormatContainersYAMLEmpty(t *testing.T) {
	expected := `docker:
  containers:
    []
`

	actual := FormatContainersYAML(nil)
	if actual != expected {
		t.Fatalf("expected:\n%s\nactual:\n%s", expected, actual)
	}
}

func TestContainerNames(t *testing.T) {
	containers := []DockerContainer{
		{Name: "nginx"},
		{Name: ""},
		{Name: "postgres"},
	}

	names := ContainerNames(containers)

	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}

	if names[0] != "nginx" {
		t.Fatalf("expected nginx, got %s", names[0])
	}

	if names[1] != "postgres" {
		t.Fatalf("expected postgres, got %s", names[1])
	}
}
