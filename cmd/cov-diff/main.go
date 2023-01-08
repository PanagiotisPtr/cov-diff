package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"golang.org/x/tools/cover"
)

var packageName = flag.String("module", "", "the name of the go module")

type FileChanges struct {
	Filename    string
	LineNumbers []int
}

func ComputeFileChangesFromHunk(
	f *diff.FileDiff,
) FileChanges {
	fc := FileChanges{}
	parts := strings.Split(f.NewName, "/")
	filename := strings.Join(parts[1:], "/")

	fc.Filename = strings.Join([]string{*packageName, filename}, "/")

	for _, h := range f.Hunks {
		lines := strings.Split(string(h.Body), "\n")
		ln := int(h.NewStartLine)
		for _, l := range lines {
			if len(l) > 0 && l[0] == '-' {
				continue
			}
			if len(l) > 0 && l[0] == '+' {
				fc.LineNumbers = append(fc.LineNumbers, ln)
			}
			ln++
		}
	}

	return fc
}

func main() {
	flag.Parse()
	fmt.Println("hello world")
	b, err := ioutil.ReadFile("testcases/1.txt")
	if err != err {
		panic(err)
	}
	fs, err := diff.ParseMultiFileDiff(b)
	if err != nil {
		panic(err)
	}

	for _, f := range fs {
		fc := ComputeFileChangesFromHunk(f)
		fmt.Println("filename: ", fc.Filename)
		fmt.Println("Lines Changed: ", fc.LineNumbers)

		_, err := cover.ParseProfiles("testcases/1.coverage")
		if err != nil {
			panic(err)
		}

		// todo
		// for _, cp := range cps {
		// }
	}
}
