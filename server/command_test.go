package main

import "testing"

func TestParseTwinmateCommand(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"/twinmate", "status"}, {"/twinmate on", "on"}, {"/twinmate off", "off"}, {"/twinmate status", "status"},
		{"/twinmate delete all", "invalid"}, {"/twinmate unknown", "invalid"},
	} {
		if got := parseTwinmateCommand(tc.input); got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.input, got, tc.want)
		}
	}
}
