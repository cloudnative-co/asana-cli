package auth

import "testing"

func TestEnvForKind(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		KindPAT:          "ASANA_CLI_PAT",
		KindAccessToken:  "ASANA_CLI_ACCESS_TOKEN",
		KindRefreshToken: "ASANA_CLI_REFRESH_TOKEN",
		KindClientSecret: "ASANA_CLI_CLIENT_SECRET",
		"unknown":        "",
	}

	for kind, want := range cases {
		if got := envForKind(kind); got != want {
			t.Fatalf("envForKind(%q) = %q, want %q", kind, got, want)
		}
	}
}
