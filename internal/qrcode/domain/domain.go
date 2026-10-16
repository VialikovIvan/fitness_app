package domain

import "errors"

var ErrUserIsInCooldown = errors.New("user is in cooldown")
