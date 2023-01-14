package diff

import (
	"strings"

	"github.com/panagiotisptr/cov-diff/files"
	"github.com/panagiotisptr/cov-diff/interval"
	godiff "github.com/sourcegraph/go-diff/diff"
)

func GetFilesIntervalsFromDiff(
	diffBytes []byte,
) (interval.FilesIntervals, error) {
	filesIntervals := interval.FilesIntervals{}

	fs, err := godiff.ParseMultiFileDiff(diffBytes)
	if err != nil {
		return filesIntervals, err
	}

	for _, f := range fs {
		parts := strings.Split(f.NewName, "/")
		filename := strings.Join(parts[1:], "/")

		if files.ShouldSkipFile(filename) {
			continue
		}

		if _, ok := filesIntervals[filename]; !ok {
			filesIntervals[filename] = []interval.Interval{}
		}
		for _, h := range f.Hunks {
			lines := strings.Split(string(h.Body), "\n")
			ln := int(h.NewStartLine)
			blockStart := 0
			inBlock := false
			for _, l := range lines {
				if len(l) > 0 && l[0] == '-' {
					continue
				}
				if len(l) > 0 && l[0] == '+' {
					if !inBlock {
						inBlock = true
						blockStart = ln
					}
					ln++
					continue
				}
				if inBlock {
					inBlock = false
					filesIntervals[filename] = append(filesIntervals[filename], interval.Interval{
						Start: blockStart,
						End:   ln,
					})
				}
				ln++
			}
			if inBlock {
				inBlock = false
				filesIntervals[filename] = append(filesIntervals[filename], interval.Interval{
					Start: blockStart,
					End:   ln - 1,
				})
			}
		}
	}

	return filesIntervals, nil
}
