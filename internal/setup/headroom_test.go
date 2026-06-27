package setup

import (
	"strings"
	"testing"
)

func TestResolveHeadroomInstall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		hasPipx      bool
		hasPip3      bool
		hasPip       bool
		wantRunnable bool
		wantName     string
	}{
		{
			name:         "prefers pipx when available",
			hasPipx:      true,
			hasPip3:      true,
			hasPip:       true,
			wantRunnable: true,
			wantName:     "pipx",
		},
		{
			name:         "falls back to pip3 without pipx",
			hasPip3:      true,
			hasPip:       true,
			wantRunnable: true,
			wantName:     "pip3",
		},
		{
			name:         "falls back to pip when only pip exists",
			hasPip:       true,
			wantRunnable: true,
			wantName:     "pip",
		},
		{
			name:         "no python installer is not runnable",
			wantRunnable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ResolveHeadroomInstall(tt.hasPipx, tt.hasPip3, tt.hasPip)
			if got.Runnable != tt.wantRunnable {
				t.Fatalf("Runnable = %t, want %t (cmd=%#v)", got.Runnable, tt.wantRunnable, got)
			}
			if !tt.wantRunnable {
				if strings.TrimSpace(got.Manual) == "" {
					t.Fatalf("expected non-runnable command to carry manual guidance, got %#v", got)
				}
				return
			}
			if got.Name != tt.wantName {
				t.Fatalf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if !strings.Contains(strings.Join(got.Args, " "), headroomPackage) {
				t.Fatalf("args %v missing %q", got.Args, headroomPackage)
			}
			if strings.TrimSpace(got.Display) == "" {
				t.Fatal("expected runnable command to carry a display string")
			}
		})
	}
}
