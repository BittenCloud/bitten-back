package dto

// GenerateUserKeyResult holds the result of generating a key for a user.
type GenerateUserKeyResult struct {
	VlessKey              string
	HasActiveSubscription bool
}
