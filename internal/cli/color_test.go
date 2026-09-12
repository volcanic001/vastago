package cli

import (
	"testing"

	"github.com/charmbracelet/colorprofile"
)

func TestTUIColorProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile colorprofile.Profile
		env     []string
		want    colorprofile.Profile
	}{
		{
			name:    "remote ANSI256 terminal keeps RGB colors",
			profile: colorprofile.ANSI256,
			env:     []string{"TERM=xterm-256color", "SSH_CONNECTION=client server"},
			want:    colorprofile.TrueColor,
		},
		{
			name:    "remote ANSI terminal keeps RGB colors",
			profile: colorprofile.ANSI,
			env:     []string{"TERM=xterm", "SSH_TTY=/dev/pts/0"},
			want:    colorprofile.TrueColor,
		},
		{
			name:    "local profile remains detected profile",
			profile: colorprofile.ANSI256,
			env:     []string{"TERM=xterm-256color"},
			want:    colorprofile.ANSI256,
		},
		{
			name:    "remote no-color preference is preserved",
			profile: colorprofile.ASCII,
			env:     []string{"TERM=xterm-256color", "SSH_CLIENT=client"},
			want:    colorprofile.ASCII,
		},
		{
			name:    "remote non-terminal output is preserved",
			profile: colorprofile.NoTTY,
			env:     []string{"TERM=xterm-256color", "SSH_CONNECTION=client server"},
			want:    colorprofile.NoTTY,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := tuiColorProfile(test.profile, test.env); got != test.want {
				t.Fatalf("tuiColorProfile(%s) = %s, want %s", test.profile, got, test.want)
			}
		})
	}
}
