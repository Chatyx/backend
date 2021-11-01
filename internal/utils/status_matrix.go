package utils

type StatusSet map[int]bool

func NewStatusSet(statuses ...int) StatusSet {
	set := make(StatusSet, len(statuses))
	for _, status := range statuses {
		set[status] = true
	}

	return set
}

type StatusMatrix map[int]StatusSet

func (mx StatusMatrix) IsCorrectTransit(from, to int) bool {
	set, ok := mx[from]
	if !ok {
		return false
	}

	return set[to]
}
