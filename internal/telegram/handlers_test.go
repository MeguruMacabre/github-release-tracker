package telegram

import "testing"

func TestParseRepository(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"/add golang/go", "golang/go", false},
		{"/add CLI/CLI", "cli/cli", false},
		{"/remove owner/repo", "owner/repo", false},
		{"/add", "", true},
		{"/add invalid", "", true},
		{"/add owner/repo/extra", "", true},
		{"/add owner/repo extra", "", true},
	}

	for _, tt := range tests {
		got, err := parseRepository(tt.input)

		if (err != nil) != tt.wantErr {
			t.Errorf("parseRepository(%q): error = %v", tt.input, err)
			continue
		}

		if got != tt.want {
			t.Errorf(
				"parseRepository(%q) = %q, want %q",
				tt.input,
				got,
				tt.want,
			)
		}
	}
}
