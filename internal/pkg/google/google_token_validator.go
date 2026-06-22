package google

import (
	"context"

	"google.golang.org/api/idtoken"
)

type GoogleTokenValidator struct{}

func NewGoogleTokenValidator() *GoogleTokenValidator {
	return &GoogleTokenValidator{}
}

func (v *GoogleTokenValidator) Validate(ctx context.Context, token string, audience string) (*idtoken.Payload, error) {
	return idtoken.Validate(
		ctx,
		token,
		audience,
	)
}
