package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/actions-go/toolkit/core"
	"github.com/panagiotisptr/cov-diff/cov"
	"github.com/panagiotisptr/cov-diff/diff"
	"github.com/panagiotisptr/cov-diff/files"
	"github.com/panagiotisptr/cov-diff/interval"
)

var path = flag.String("path", "", "path to the git repository")
var coverageFile = flag.String("coverprofile", "", "location of the coverage file")
var sourceBranch = flag.String("source", "", "the name of the source branch (the one we have coverage for)")
var targetBranch = flag.String("target", "", "the name of the target branch (usually main/master)")
var moduleName = flag.String("module", "", "the name of module")

func main() {
	core.Debug("Running action")
	core.SetOutput("myOutput", fmt.Sprintf("Hello %s", os.Getenv("INPUT_MYINPUT")))

	os.Exit(0)

	flag.Parse()
	if *coverageFile == "" {
		log.Fatal("missing coverage file")
	}

	diffBytes, err := exec.Command(
		"sh",
		"-c",
		fmt.Sprintf(
			"(cd %s && git diff %s %s)",
			*path,
			*targetBranch,
			*sourceBranch,
		),
	).Output()
	if err != nil {
		log.Fatal(err)
	}

	diffIntervals, err := diff.GetFilesIntervalsFromDiff(diffBytes)
	if err != nil {
		log.Fatal(err)
	}
	// de-allocate diffBytes
	diffBytes = nil

	covFileBytes, err := os.ReadFile(*coverageFile)
	if err != nil {
		log.Fatal(err)
	}

	coverIntervals, err := cov.GetFilesIntervalsFromCoverage(covFileBytes)
	if err != nil {
		log.Fatal(err)
	}
	// de-allocate covFileBytes
	covFileBytes = nil

	total := 0
	covered := 0
	for filename, di := range diffIntervals {
		fileBytes, err := os.ReadFile(filepath.Join(*path, filename))
		if err != nil {
			log.Fatal(err)
		}
		fi, err := files.GetIntervalsFromFile(fileBytes)
		if err != nil {
			log.Fatal(err)
		}

		// intervals that changed and are parts of the code we care about
		measuredIntervals := interval.Union(di, fi)
		total += interval.Sum(measuredIntervals)

		fullFilename := filepath.Join(*moduleName, filename)
		ci, ok := coverIntervals[fullFilename]
		if !ok {
			continue
		}

		coveredMeasuredIntervals := interval.Union(measuredIntervals, ci)
		covered += interval.Sum(coveredMeasuredIntervals)
	}

	percentCoverage := 100
	if total != 0 {
		percentCoverage = (100 * covered) / total
	}

	fmt.Printf("Coverage on new lines: %d%%\n", percentCoverage)
}
