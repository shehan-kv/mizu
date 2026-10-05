package message

type Stats struct {
	fileCount int
}

func NewStats(fileCount int) Stats {
	return Stats{fileCount: fileCount}
}

func (s Stats) FileCount() int {
	return s.fileCount
}
