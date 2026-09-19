package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"ph-province/internal/boundary"
	"ph-province/internal/render"
	"ph-province/internal/terminal"
)

const usage = `Usage:
  ph-province             Start the interactive prompt
  ph-province PROVINCE    Render a province and exit
  ph-province --list      List all supported provinces
  ph-province --help      Show this help
  ph-province --version   Show the version

Examples:
  ph-province cebu
  ph-province agusan del norte
`

type runtime struct {
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
	version string
	size    func() (int, int)
}

// Run executes the CLI and returns its process exit code.
func Run(args []string, stdin *os.File, stdout *os.File, stderr io.Writer, version string) int {
	runtime := runtime{
		stdin: stdin, stdout: stdout, stderr: stderr, version: version,
		size: func() (int, int) { return terminal.Size(stdout) },
	}
	return runtime.run(args)
}

func (runtime runtime) run(args []string) int {
	switch {
	case len(args) == 0:
		return runtime.interactive()
	case len(args) == 1 && (args[0] == "--help" || args[0] == "-h"):
		fmt.Fprint(runtime.stdout, usage)
		return 0
	case len(args) == 1 && args[0] == "--version":
		fmt.Fprintf(runtime.stdout, "ph-province %s\n", runtime.version)
		return 0
	case len(args) == 1 && args[0] == "--list":
		if err := runtime.printList(); err != nil {
			fmt.Fprintf(runtime.stderr, "ph-province: %v\n", err)
			return 1
		}
		return 0
	case len(args) >= 1 && !strings.HasPrefix(args[0], "-"):
		if err := runtime.draw(strings.Join(args, " "), false); err != nil {
			fmt.Fprintf(runtime.stderr, "ph-province: %v\n", err)
			return 1
		}
		return 0
	default:
		fmt.Fprintln(runtime.stderr, "ph-province: invalid arguments")
		fmt.Fprint(runtime.stderr, usage)
		return 2
	}
}

func (runtime runtime) interactive() int {
	scanner := bufio.NewScanner(runtime.stdin)
	for {
		fmt.Fprint(runtime.stdout, "province> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(runtime.stderr, "ph-province: read input: %v\n", err)
				return 1
			}
			return 0
		}
		province := strings.TrimSpace(scanner.Text())
		if province == "" {
			continue
		}
		if strings.EqualFold(province, "quit") || strings.EqualFold(province, "exit") {
			return 0
		}
		if strings.EqualFold(province, "list") {
			if err := runtime.printList(); err != nil {
				fmt.Fprintf(runtime.stderr, "ph-province: %v\n", err)
			}
			continue
		}
		if err := runtime.draw(province, true); err != nil {
			fmt.Fprintf(runtime.stderr, "ph-province: %v\n", err)
		}
	}
}

func (runtime runtime) draw(province string, interactive bool) error {
	match, err := boundary.Resolve(province)
	if err != nil {
		return err
	}
	columns, rows := runtime.size()
	if interactive {
		rows = max(2, rows-2)
	}
	output, err := render.Draw(match.Shape, columns, rows)
	if err != nil {
		return err
	}
	_, err = io.WriteString(runtime.stdout, output)
	return err
}

func (runtime runtime) printList() error {
	names, err := boundary.Names()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(runtime.stdout, strings.Join(names, "\n"))
	return err
}
