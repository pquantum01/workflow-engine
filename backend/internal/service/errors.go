package service

import "errors"

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
var ErrInstanceTerminal = errors.New("instance is in terminal state")
var ErrTaskNotCompletable = errors.New("task is not in a completable state")
var ErrInvalidNextState = errors.New("next_state is not a valid outgoing transition")
var ErrDefHasActiveInstances = errors.New("workflow def has active instances")
var ErrValidation        = errors.New("validation error")
var ErrTaskCompleted     = errors.New("task is completed, field update not allowed")
var ErrInvalidState      = errors.New("state_id does not exist in workflow definition")
var ErrInstanceNotActive = errors.New("instance is not active")
