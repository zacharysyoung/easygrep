package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(t *testing.T) {
	testCases := []struct {
		testname string
		args     []string

		code     int
		out, err string
	}{
		{
			"single file",
			[]string{"foo", "testdata/a.txt"},
			0,
			`
			testdata/a.txt:2:foo
			`,
			``,
		},
		{
			"multiple files",
			[]string{"foo", "testdata/1.txt", "testdata/a.txt"},
			0,
			`
			testdata/1.txt:3:foo
			testdata/a.txt:2:foo
			`,
			``,
		},
		{
			"bad file",
			[]string{"foo", "testdata/1.txt", "testdata/z.txt"},
			0,
			`
			testdata/1.txt:3:foo
			`,
			`
			grep: open testdata/z.txt: no such file or directory
			`,
		},
		{
			"recurse directories",
			[]string{"foo", "testdata/c", "testdata/d"},
			0,
			`
			testdata/c/c.txt:1:foo
			testdata/d/d/d.txt:3:foo
			`,
			``,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testname, func(t *testing.T) {
			outbuf, errbuf := &bytes.Buffer{}, &bytes.Buffer{}

			code := Main(outbuf, errbuf, tc.args...)
			if code != tc.code {
				t.Fatalf("got code %d; want %d", code, tc.code)
			}

			got := outbuf.String()
			want := preprocess(tc.out)
			if got != want {
				t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
			}

			got = errbuf.String()
			want = preprocess(tc.err)
			if got != want {
				t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
			}
		})
	}

	t.Run("absolute path", func(t *testing.T) {
		tempdir := t.TempDir()
		path := filepath.Join(tempdir, "a.txt")
		f, _ := os.Create(path)
		io.WriteString(f, "foo\n")
		f.Close()

		outbuf, errbuf := &bytes.Buffer{}, &bytes.Buffer{}

		code := Main(outbuf, errbuf, "foo", path)
		if code != 0 {
			t.Fatalf("got code %d; want 0", code)
		}

		got := outbuf.String()
		want := preprocess(path + ":1:foo\n")
		if got != want {
			t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
		}

		got = errbuf.String()
		want = ""
		if got != want {
			t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
		}
	})
}

func preprocess(s string) string {
	s = strings.TrimLeft(s, " \n")
	s = strings.ReplaceAll(s, "\t", "")
	return s
}
