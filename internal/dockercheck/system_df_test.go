package dockercheck

import (
	"testing"

	"github.com/IKHINtech/composeguard/internal/checker"
	"github.com/IKHINtech/composeguard/internal/config"
)

func TestParseDockerSizeToBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint64
	}{
		{
			name:     "bytes",
			input:    "12B",
			expected: 12,
		},
		{
			name:     "kilobytes decimal",
			input:    "1kB",
			expected: 1000,
		},
		{
			name:     "megabytes decimal",
			input:    "1.5MB",
			expected: 1500000,
		},
		{
			name:     "gigabytes decimal",
			input:    "2GB",
			expected: 2000000000,
		},
		{
			name:     "gibibytes binary",
			input:    "1GiB",
			expected: 1073741824,
		},
		{
			name:     "with space",
			input:    "10.5 GB",
			expected: 10500000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseDockerSizeToBytes(tt.input)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if actual != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, actual)
			}
		})
	}
}

func TestParseDockerSizeToBytesInvalid(t *testing.T) {
	tests := []string{
		"",
		"abc",
		"12XB",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := parseDockerSizeToBytes(input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestCheckSystemDF(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		cfg  config.DockerSystemDFConfig
		want []checker.Result
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckSystemDF(tt.cfg)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("CheckSystemDF() = %v, want %v", got, tt.want)
			}
		})
	}
}
