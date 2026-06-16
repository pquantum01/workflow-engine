package service

import "errors"

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
var ErrInstanceTerminal = errors.New("instance is in terminal state")
var ErrTaskNotCompletable = errors.New("task is not in a completable state")
var ErrInvalidNextState = errors.New("next_state is not a valid outgoing transition")
var ErrDefHasActiveInstances = errors.New("workflow def has active instances")
var ErrValidation = errors.New("validation error")
