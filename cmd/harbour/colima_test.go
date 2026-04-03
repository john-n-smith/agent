package main

import (
	"reflect"
	"testing"
)

func TestFormatMountDiff(t *testing.T) {
	tests := []struct {
		name    string
		current []string
		desired []string
		want    []string
	}{
		{
			name:    "no changes",
			current: []string{"/workspace|rw"},
			desired: []string{"/workspace|rw"},
			want:    nil,
		},
		{
			name:    "added mount",
			current: []string{"/workspace|rw"},
			desired: []string{"/repo|rw", "/workspace|rw"},
			want:    []string{"+ /repo (rw)"},
		},
		{
			name:    "removed mount",
			current: []string{"/repo|rw", "/workspace|rw"},
			desired: []string{"/workspace|rw"},
			want:    []string{"- /repo (rw)"},
		},
		{
			name:    "mode change",
			current: []string{"/repo|ro"},
			desired: []string{"/repo|rw"},
			want:    []string{"- /repo (ro)", "+ /repo (rw)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMountDiff(tt.current, tt.desired)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("formatMountDiff() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestHumanizeMountLine(t *testing.T) {
	got := humanizeMountLine("/repo|rw")
	want := "/repo (rw)"
	if got != want {
		t.Fatalf("humanizeMountLine() = %q, want %q", got, want)
	}
}

func TestNormalizeMountLines(t *testing.T) {
	got := normalizeMountLines([]string{
		"/workspace|rw",
		"/repo|rw",
		"/workspace|rw",
		"/repo|rw",
	})
	want := []string{
		"/repo|rw",
		"/workspace|rw",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeMountLines() = %#v, want %#v", got, want)
	}
}

func TestFormatMountDiffIgnoresDuplicateMountLines(t *testing.T) {
	current := normalizeMountLines([]string{
		"/workspace|rw",
		"/workspace|rw",
	})
	desired := normalizeMountLines([]string{
		"/workspace|rw",
	})

	got := formatMountDiff(current, desired)
	if got != nil {
		t.Fatalf("formatMountDiff() = %#v, want nil", got)
	}
}
