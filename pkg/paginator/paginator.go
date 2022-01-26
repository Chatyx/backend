package paginator

type Paginate interface {
	HasNext() bool
	HasPrev() bool
}
