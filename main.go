package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"regexp/syntax"
)

func usage() {
	fmt.Fprintln(os.Stderr, ""+
		`usage: grep [-i] pattern [file...]

If grep does not find pattern in any of the files it exits with status code 1.`)
	flag.PrintDefaults()
	os.Exit(2)
}

var (
	iflag = flag.Bool("i", false, "Perform case insensitive matching.")

	rePattern *regexp.Regexp
)

func main() {
	flag.Usage = usage
	flag.Parse()
	code := Main(os.Stdout, os.Stderr, flag.Args()...)
	os.Exit(code)
}

// Main takes the stdout and stderr writers, and the args list
// (which should include the pattern first), and runs the matcher
// before finally returning a 0 for some match found or 1 for
// no match.
func Main(stdout, stderr io.Writer, args ...string) (retcode int) {
	if len(args) < 2 {
		usage()
	}

	processPattern(args[0])

	matched := false
	for _, path := range args[1:] {
		switch isDir(path) {
		case true:
			for _, path := range walkDir(path) {
				if match(path, stdout, stderr) {
					matched = true
				}
			}
		case false:
			if match(path, stdout, stderr) {
				matched = true
			}
		}
	}

	if matched {
		return 0
	}
	return 1
}

// walkDir walks dir, which can be a relative or absolute path.
func walkDir(dir string) (paths []string) {
	fsys := os.DirFS(".")
	prefix := ""
	if filepath.IsAbs(dir) {
		fsys = os.DirFS("/")
		dir = dir[1:]
		prefix = "/"
	}

	fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		path = prefix + path
		paths = append(paths, path)
		return nil
	})

	return paths
}

// isDir reports whether path is a directory.
func isDir(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// match writes a match to stdout, errors to stderr, and
// finally if a match was found.
func match(path string, stdout, stderr io.Writer) (matched bool) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(stderr, "grep: %v\n", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for i := 1; scanner.Scan(); i++ {
		line := scanner.Bytes()
		if rePattern.Match(line) {
			matched = true
			fmt.Fprintf(stdout, "%s:%d:%s\n", path, i, line)
		}
	}
	return matched
}

// processPattern creates the global regexp from pattern.
func processPattern(pattern string) {
	if *iflag {
		pattern = "(?i)" + pattern
	}

	var err error
	rePattern, err = regexp.Compile(pattern)
	if err != nil {
		err := err.(*syntax.Error)
		fatalf("%s", err.Code)
	}
}

func fatalf(format string, args ...any) {
	format = "grep: " + format + "\n"
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(2)
}
