package files

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/panagiotisptr/cov-diff/interval"
)

var LineNotFound error = errors.New("line not found")

func ShouldSkipFile(filename string) bool {
	if strings.Contains(filename, "_test.go") {
		return true
	}
	if len(filename) > 3 && filename[len(filename)-3:] != ".go" {
		return true
	}
	if strings.Contains(filename, "vendor/") {
		return true
	}

	return false
}

type line struct {
	Start  token.Pos
	End    token.Pos
	Number int
}

func GetLineFromToken(
	lines []line,
	pos token.Pos,
) (int, error) {
	l := lines
	for {
		if len(l) == 0 {
			return 0, LineNotFound
		}
		mid := len(l) / 2
		if l[mid].Start <= pos && l[mid].End >= pos {
			return l[mid].Number, nil
		} else if l[mid].End < pos {
			l = l[mid+1:]
		} else {
			l = l[:mid]
		}
	}
}

func GetIntervalsFromFile(
	fileBytes []byte,
) ([]interval.Interval, error) {
	intervals := []interval.Interval{}
	fileLines := strings.Split(string(fileBytes), "\n")
	count := 0
	// this will be sorted
	lines := make([]line, len(fileLines)+1)
	for i, fl := range fileLines {
		lines[i+1] = line{
			Start:  token.Pos(count),
			End:    token.Pos(count + len(fl)),
			Number: i + 1,
		}
		count += len(fl) + 1
	}

	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, "", fileBytes, 0)
	if err != nil {
		return intervals, err
	}

	// we don't care about coverage on the main package
	if parsedFile.Name.Name == "main" {
		return intervals, err
	}

	for _, d := range parsedFile.Decls {
		switch mid := d.(type) {
		case *ast.FuncDecl:
			if mid.Body.Pos().IsValid() && mid.Body.End().IsValid() {
				startLine, err := GetLineFromToken(lines, mid.Body.Pos())
				if err != nil {
					return intervals, err
				}
				endLine, err := GetLineFromToken(lines, mid.Body.End()-1)
				if err != nil {
					return intervals, err
				}
				intervals = append(intervals, interval.Interval{
					Start: startLine,
					End:   endLine,
				})
			}
		}
	}

	return intervals, nil
}
