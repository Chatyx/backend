package dto

type DialogCreate struct {
	PartnerUserID int
}

type DialogUpdate struct {
	ID               int
	PartnerIsBlocked *bool
}
