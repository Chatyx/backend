package paginator

type LimitOffsetPaginate struct {
	Total  int
	Count  int
	Limit  int
	Offset int
}

func (p LimitOffsetPaginate) HasNext() bool {
	return p.Total > 0 && p.Offset+p.Count < p.Total
}

func (p LimitOffsetPaginate) HasPrev() bool {
	return p.Total > 0 && p.Offset > 0
}
