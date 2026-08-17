package gitutil

import "testing"

func TestExtractTicket(t *testing.T) {
	const pattern = `[A-Z]{2,}-\d+`
	cases := map[string]string{
		"feature/VAL-482-refonte-auth": "VAL-482",
		"fix/PROJ-12":                  "PROJ-12",
		"main":                         "",
		"chore/nettoyage":              "",
		"BOTERO-9001-mercure":          "BOTERO-9001",
	}
	for branch, want := range cases {
		if got := ExtractTicket(branch, pattern); got != want {
			t.Errorf("ExtractTicket(%q) = %q, attendu %q", branch, got, want)
		}
	}
}
