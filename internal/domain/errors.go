package domain

import "errors"

var (
	ErrInvalidNumberOfChoice = errors.New("invalid number of choices for a multiple choice question")
	ErrNoCorrectChoice       = errors.New("no correct choice")
	ErrQuestionNotFound      = errors.New("question not found")
	ErrModuleNotFound        = errors.New("module not found")
	ErrInvalidData           = errors.New("invalid data")
	ErrRecordNotFound        = errors.New("record not found")
	ErrDuplicateEmail        = errors.New("this email is already taken")
	ErrForeignConstraint     = errors.New("foreign key constraint violated")
	ErrConflict              = errors.New("conflict request")

	ErrInvalidChoice = errors.New("invalid choice id. Choice not found.")

	ErrFriendsAlreadyExist = errors.New("this friendship already exists")
	ErrSelfCantBeFriend    = errors.New("you can't be friends with yourself")
)
