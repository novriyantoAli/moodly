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

func HasPermission(ctx context.Context, permission string) bool {
	p, ok := PrincipalFromContext(ctx)
	if !ok {
		return false
	}
	for _, perm := range p.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

func HasRole(ctx context.Context, role string) bool {
	p, ok := PrincipalFromContext(ctx)
	if !ok {
		return false
	}
	for _, r := range p.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func HasAnyIntersectionRoles(ctx context.Context, roles []string) bool {
	p, ok := PrincipalFromContext(ctx)
	if !ok {
		return false
	}

	roles1 := p.Roles
	roles2 := roles

	if len(roles1) > len(roles2) {
		roles1, roles2 = roles2, roles1
	}

	set := make(map[string]struct{}, len(roles1))
	for _, v := range roles1 {
		set[v] = struct{}{}
	}

	for _, v := range roles2 {
		if _, exists := set[v]; exists {
			return true
		}
	}

	return false
}