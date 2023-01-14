package interval

type Interval struct {
	Start int
	End   int
}

type FilesIntervals = map[string][]Interval
