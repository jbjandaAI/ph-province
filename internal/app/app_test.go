package app

import (
	"bytes"
	"strings"
	"testing"
)

func testRuntime(input string) (runtime, *bytes.Buffer, *bytes.Buffer) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	return runtime{
		stdin: strings.NewReader(input), stdout: stdout, stderr: stderr,
		version: "test", size: func() (int, int) { return 40, 24 },
	}, stdout, stderr
}

func TestDirectCebu(t *testing.T) {
	runtime, stdout, stderr := testRuntime("")
	if code := runtime.run([]string{"  CeBu  "}); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.ContainsAny(stdout.String(), "▀▄█") {
		t.Fatalf("output does not contain a rendered shape: %q", stdout.String())
	}
}

func TestUnknownProvince(t *testing.T) {
	runtime, _, stderr := testRuntime("")
	if code := runtime.run([]string{"bohol"}); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "supports: cebu") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestInteractive(t *testing.T) {
	runtime, stdout, stderr := testRuntime("\nCEBU\nbohol\nquit\n")
	if code := runtime.run(nil); code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if got := strings.Count(stdout.String(), "province> "); got != 4 {
		t.Fatalf("prompt count = %d, want 4", got)
	}
	if !strings.Contains(stderr.String(), "unknown province") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestHelpAndVersion(t *testing.T) {
	for _, test := range []struct {
		args []string
		want string
	}{
		{args: []string{"--help"}, want: "ph-province cebu"},
		{args: []string{"--version"}, want: "ph-province test"},
	} {
		runtime, stdout, _ := testRuntime("")
		if code := runtime.run(test.args); code != 0 {
			t.Fatalf("%v exit code = %d", test.args, code)
		}
		if !strings.Contains(stdout.String(), test.want) {
			t.Fatalf("%v output = %q", test.args, stdout.String())
		}
	}
}
