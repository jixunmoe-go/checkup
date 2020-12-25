package types

type HistoryReader interface {
	// Fetch returns the contents of a check file.
	Fetch(checkFile string) ([]Result, error)
	// GetIndex returns the storage index, as a map where keys are check
	// result filenames and values are the associated check timestamps.
	GetIndex() (map[string]int64, error)

	FindLastWorkingTime(current *Result) int64
}

type DownTimeNotifier interface {
	Type() string
	NotifyDowntime(current, prev []Result, reader HistoryReader) error
}
