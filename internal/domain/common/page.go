package common

type Page struct {
	limit  int
	offset int
}

func NewPage(limit, offset int) (Page, error) {
	if limit <= 0 {
		return Page{}, ErrInvalidPageLimit
	}
	if limit > 200 {
		return Page{}, ErrPageLimitExceeded
	}
	if offset < 0 {
		return Page{}, ErrInvalidPageOffset
	}
	return Page{limit: limit, offset: offset}, nil
}

func (p Page) Limit() int  { return p.limit }
func (p Page) Offset() int { return p.offset }
