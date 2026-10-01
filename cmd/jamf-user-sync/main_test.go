package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestInformationalFlagsDoNotRequireCredentials(t *testing.T) {
	for _, flag := range []string{"--help", "--version"} {
		var stdout, stderr bytes.Buffer
		if code := run(t.Context(), []string{flag}, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "jamf-user-sync") || stderr.Len() != 0 {
			t.Errorf("run(%s) = %d, stdout %q, stderr %q", flag, code, stdout.String(), stderr.String())
		}
	}
}

func TestUnexpectedArgumentsFailBeforeConnecting(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(t.Context(), []string{"serve"}, &stdout, &stderr); code != 2 {
		t.Fatalf("unexpected argument exit = %d", code)
	}
}
