package services

import "github.com/google/uuid"

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

// FreeTierUserUUID is a predefined UUID for users accessing free tier keys without registration.
var FreeTierUserUUID = uuid.MustParse("5ccc43c4-3c3e-4220-a878-761aa1182dd9")
