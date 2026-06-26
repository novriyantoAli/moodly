package security

import "context"

type contextKey string

const principalKey contextKey = "principal"

func WithPrincipal(
	ctx context.Context,
	p Principal,
) context.Context {
	return context.WithValue(
		ctx,
		principalKey,
		p,
	)
}

func PrincipalFromContext(
	ctx context.Context,
) (Principal, bool) {
	p, ok := ctx.Value(principalKey).(Principal)
	return p, ok
}

func MustPrincipal(
	ctx context.Context,
) Principal {
	p, ok := PrincipalFromContext(ctx)
	if !ok {
		panic("principal not found")
	}

	return p
}