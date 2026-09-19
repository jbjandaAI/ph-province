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

func TestDirectMultiwordProvince(t *testing.T) {
	runtime, stdout, stderr := testRuntime("")
	if code := runtime.run([]string{"Agusan", "del", "Norte"}); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.ContainsAny(stdout.String(), "▀▄█") {
		t.Fatalf("output does not contain a rendered shape: %q", stdout.String())
	}
}

func TestUnknownProvinceSuggestsMatch(t *testing.T) {
	runtime, _, stderr := testRuntime("")
	if code := runtime.run([]string{"boholl"}); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "did you mean: Bohol") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestAmbiguousMaguindanao(t *testing.T) {
	runtime, _, stderr := testRuntime("")
	if code := runtime.run([]string{"Maguindanao"}); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "Maguindanao del Norte, Maguindanao del Sur") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestList(t *testing.T) {
	runtime, stdout, stderr := testRuntime("")
	if code := runtime.run([]string{"--list"}); code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 82 {
		t.Fatalf("listed %d provinces, want 82", len(lines))
	}
	if lines[0] != "Abra" || lines[len(lines)-1] != "Zamboanga Sibugay" {
		t.Fatalf("unexpected list bounds: first=%q last=%q", lines[0], lines[len(lines)-1])
	}
}

func TestInteractive(t *testing.T) {
	runtime, stdout, stderr := testRuntime("\nlist\nCEBU\nboholl\nquit\n")
	if code := runtime.run(nil); code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if got := strings.Count(stdout.String(), "province> "); got != 5 {
		t.Fatalf("prompt count = %d, want 5", got)
	}
	if !strings.Contains(stdout.String(), "Zamboanga Sibugay\n") {
		t.Fatal("interactive list did not print all province names")
	}
	if !strings.Contains(stderr.String(), "did you mean: Bohol") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestHelpAndVersion(t *testing.T) {
	for _, test := range []struct {
		args []string
		want string
	}{
		{args: []string{"--help"}, want: "ph-province --list"},
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
