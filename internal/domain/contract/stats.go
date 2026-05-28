package contract

type Stats struct {
	total  int
	signed int
}

func NewStats(total int, signed int) Stats {
	return Stats{total: total, signed: signed}
}

func (s Stats) Total() int {
	return s.total
}

func (s Stats) Signed() int {
	return s.signed
}
