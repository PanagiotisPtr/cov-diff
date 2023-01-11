package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"golang.org/x/tools/cover"
)

var moduleName = flag.String("module", "", "the name of the go module")
var path = flag.String("path", "", "the path to the module")
var coverageFile = flag.String("coverfile", "", "location of the coverage file")
var sourceBranch = flag.String("source", "", "the name of the source branch (the one we have coverage for)")
var targetBranch = flag.String("target", "", "the name of the target branch (usually main/master)")

type FileChanges struct {
	Filename    string
	LineNumbers []int
}

type AreaOfInterest struct {
	Start token.Pos
	End   token.Pos
}

func ComputeFileChangesFromHunk(
	f *diff.FileDiff,
) FileChanges {
	fc := FileChanges{}
	parts := strings.Split(f.NewName, "/")
	filename := strings.Join(parts[1:], "/")

	fc.Filename = filename

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

func exitWithError(err error, msg string) {
	fmt.Printf("%s: %v\n", msg, err)
	os.Exit(1)
}

func main() {
	out, err := exec.Command("git", "diff").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The date is %s\n", out)
	return
	flag.Parse()
	b, err := os.ReadFile(filepath.Join(*path, "testcases/1.txt"))
	if err != err {
		panic(err)
	}
	fs, err := diff.ParseMultiFileDiff(b)
	if err != nil {
		panic(err)
	}

	cps, err := cover.ParseProfiles(
		filepath.Join(*path, "testcases/1.coverage"),
	)
	if err != nil {
		panic(err)
	}

	fileCPs := map[string][]*cover.Profile{}
	for _, cp := range cps {
		relativeFilename := strings.Split(cp.FileName, *moduleName+"/")[1]
		if _, ok := fileCPs[relativeFilename]; !ok {
			fileCPs[relativeFilename] = []*cover.Profile{}
		}
		fileCPs[relativeFilename] = append(fileCPs[relativeFilename], cp)
	}

	totalLines := 0
	coveredLines := 0
	for _, f := range fs {
		fc := ComputeFileChangesFromHunk(f)

		if strings.Contains(fc.Filename, "_test.go") {
			continue
		}
		if len(fc.Filename) > 3 && fc.Filename[len(fc.Filename)-3:] != ".go" {
			continue
		}
		if strings.Contains(fc.Filename, "vendor/") {
			continue
		}
		aois := []AreaOfInterest{}

		fb, err := os.ReadFile(filepath.Join(*path, fc.Filename))
		if err != err {
			panic(err)
		}

		fileLines := strings.Split(string(fb), "\n")
		lineToToken := make([]AreaOfInterest, len(fileLines)+1)
		count := 0
		for i, fl := range fileLines {
			lineToToken[i+1] = AreaOfInterest{
				Start: token.Pos(count),
				End:   token.Pos(count + len(fl)),
			}
			count += len(fl) + 1
		}

		fset := token.NewFileSet()
		parsedFile, err := parser.ParseFile(fset, fc.Filename, nil, 0)
		if err != nil {
			panic(err)
		}

		// we don't care about coverage on the main package
		if parsedFile.Name.Name == "main" {
			continue
		}

		for _, d := range parsedFile.Decls {
			switch t := d.(type) {
			case *ast.FuncDecl:
				if t.Body.Pos().IsValid() && t.Body.End().IsValid() {
					aois = append(aois, AreaOfInterest{
						Start: t.Body.Pos() - 1,
						End:   t.Body.End(),
					})
				}
			}
		}

		// lines of interest
		loi := map[int]bool{}
		for _, ln := range fc.LineNumbers {
			if ln < 0 && ln >= len(lineToToken) {
				continue
			}

			pos := lineToToken[ln]
			found := false
			for _, p := range aois {
				if p.Start <= pos.Start && p.End >= pos.Start {
					found = true
					break
				}
				if p.Start <= pos.End && p.End >= pos.End {
					found = true
					break
				}
			}

			if !found {
				continue
			}

			loi[ln] = false
		}

		for range loi {
			totalLines++
		}

		if _, ok := fileCPs[fc.Filename]; !ok {
			continue
		}

		for _, cps := range fileCPs[fc.Filename] {
			for _, b := range cps.Blocks {
				for i := b.StartLine; i <= b.EndLine; i++ {
					if _, ok := loi[i]; ok {
						loi[i] = true
					}
				}
			}
		}

		for _, v := range loi {
			if v {
				coveredLines++
			}
		}
	}

	fmt.Println("coverage on new code: ", (100*coveredLines)/totalLines, "%")
}
