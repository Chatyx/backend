package entity

import "errors"

var (
	ErrUserNotFound          = errors.New("user is not found")
	ErrSuchUserAlreadyExists = errors.New("user with such username or email already exists")
	ErrWrongCurrentPassword  = errors.New("wrong current password")

	ErrGroupNotFound                          = errors.New("group is not found")
	ErrGroupParticipantNotFound               = errors.New("group participant is not found")
	ErrDialogNotFound                         = errors.New("dialog is not found")
	ErrSuchDialogAlreadyExists                = errors.New("such a dialog already exists")
	ErrCreateDialogWithYourself               = errors.New("creating a dialog with yourself")
	ErrCreateDialogWithNonExistentUser        = errors.New("creating a dialog with a non-existent user")
	ErrIncorrectGroupParticipantStatusTransit = errors.New("incorrect group participant status transit")
	ErrSuchGroupParticipantAlreadyExists      = errors.New("such a group participant already exists")
	ErrAddNonExistentUserToGroup              = errors.New("addition non-existent user to group")
	ErrForbiddenPerformAction                 = errors.New("it's forbidden to perform this action")
)
