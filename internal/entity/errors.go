package entity

import "errors"

var (
	ErrUserNotFound          = errors.New("user is not found")
	ErrSuchUserAlreadyExists = errors.New("user with such username or email already exists")
	ErrWrongCurrentPassword  = errors.New("wrong current password")

	ErrGroupNotFound                         = errors.New("group is not found")
	ErrDialogNotFound                        = errors.New("dialog is not found")
	ErrSuchDialogAlreadyExists               = errors.New("such a dialog already exists")
	ErrCreatingDialogWithYourself            = errors.New("creating a dialog with yourself")
	ErrCreatingDialogWithNonExistencePartner = errors.New("creating a dialog with a non-existent partner")
)
