package main

import "testing"

func TestBoolFlagEqualsForm(t *testing.T) {
	cases := []struct {
		args []string
		want bool
	}{
		{[]string{"--dry-run"}, true},
		{[]string{"-dry-run"}, true},
		{[]string{"--dry-run=true"}, true},
		{[]string{"--dry-run=1"}, true},
		{[]string{"--dry-run=yes"}, true},
		{[]string{"-dry-run=true"}, true},
		{[]string{"--dry-run=false"}, false},
		{[]string{"--dry-run=0"}, false},
		{[]string{"--json"}, false},
		{[]string{}, false},
	}
	for _, tc := range cases {
		got := boolFlag(tc.args, "--dry-run", "-dry-run")
		if got != tc.want {
			t.Fatalf("args=%v got=%v want=%v", tc.args, got, tc.want)
		}
	}
}

func TestStripGlobalFlagsDryRunEquals(t *testing.T) {
	got := stripGlobalFlags([]string{
		"create", "-objective", "x", "--dry-run=true", "--json", "--target=.",
	})
	want := []string{"create", "-objective", "x"}
	if len(got) != len(want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got=%v want=%v", got, want)
		}
	}
}
