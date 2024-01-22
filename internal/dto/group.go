package dto

type GroupCreate struct {
	Name        string
	Description string
}

type GroupUpdate struct {
	ID          int
	Name        string
	Description string
}
