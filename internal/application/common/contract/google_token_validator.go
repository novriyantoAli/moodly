package contract

import (
	"context"

	"google.golang.org/api/idtoken"
)

type GoogleTokenValidator interface {
	Validate(ctx context.Context, token string, audience string) (*idtoken.Payload, error)
}
