package cov

import (
	"bytes"

	"github.com/panagiotisptr/cov-diff/files"
	"github.com/panagiotisptr/cov-diff/interval"
	"golang.org/x/tools/cover"
)

func GetFilesIntervalsFromCoverage(
	covBytes []byte,
) (interval.FilesIntervals, error) {
	filesIntervals := interval.FilesIntervals{}

	cps, err := cover.ParseProfilesFromReader(bytes.NewReader(covBytes))
	if err != nil {
		return filesIntervals, err
	}

	for _, cp := range cps {
		if files.ShouldSkipFile(cp.FileName) {
			continue
		}

		if _, ok := filesIntervals[cp.FileName]; !ok {
			filesIntervals[cp.FileName] = []interval.Interval{}
		}
		for _, b := range cp.Blocks {
			filesIntervals[cp.FileName] = append(filesIntervals[cp.FileName], interval.Interval{
				Start: b.StartLine,
				End:   b.EndLine,
			})
		}
	}

	return filesIntervals, nil
}
