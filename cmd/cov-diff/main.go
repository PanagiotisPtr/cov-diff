package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

func emptyValAndActionInputSet(val string, input string) bool {
	return val == "" && os.Getenv(
		fmt.Sprintf("INPUT_%s", strings.ToUpper(input)),
	) != ""
}

func getActionInput(input string) string {
	return os.Getenv(
		fmt.Sprintf("INPUT_%s", strings.ToUpper(input)),
	)
}

func populateFlagsFromActionEnvs() {
	if emptyValAndActionInputSet(*path, "path") {
		*path = getActionInput("path")
	}
	if emptyValAndActionInputSet(*coverageFile, "coverprofile") {
		*coverageFile = getActionInput("coverprofile")
	}
	if emptyValAndActionInputSet(*sourceBranch, "source") {
		*sourceBranch = getActionInput("source")
	}
	if emptyValAndActionInputSet(*targetBranch, "target") {
		*targetBranch = getActionInput("target")
	}
	if emptyValAndActionInputSet(*moduleName, "module") {
		*moduleName = getActionInput("module")
	}
}

func main() {
	flag.Parse()
	populateFlagsFromActionEnvs()

	fmt.Println(*path)
	fmt.Println(*coverageFile)
	fmt.Println(*sourceBranch)
	fmt.Println(*targetBranch)
	fmt.Println(*moduleName)

	if *coverageFile == "" {
		log.Fatal("missing coverage file")
	}

	fmt.Printf(
		"running: sh -c (cd %s && git diff %s %s)\n",
		*path,
		*targetBranch,
		*sourceBranch,
	)

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
		log.Fatal(err, "failed to get diff")
	}

	fmt.Println("diff bytes: ", string(diffBytes))

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

	if getActionInput("coverprofile") != "" {
		_, outputErr := exec.Command(
			"sh",
			"-c",
			fmt.Sprintf(`echo "{covdiff}={%d}" >> $GITHUB_OUTPUT`, percentCoverage),
		).Output()
		if outputErr != nil {
			log.Fatal(outputErr, "failed to write output")
		}
	}
}
