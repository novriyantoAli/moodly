package domain

import "github.com/novriyantoAli/moodly/internal/application/payment/entity"

var transitions = map[entity.PaymentStatus][]entity.PaymentStatus{
	entity.PaymentStatusPending: {
		entity.PaymentStatusProcessing,
		entity.PaymentStatusCanceled,
		entity.PaymentStatusExpired,
	},

	entity.PaymentStatusProcessing: {
		entity.PaymentStatusCompleted,
		entity.PaymentStatusFailed,
		entity.PaymentStatusExpired,
	},

	entity.PaymentStatusFailed: {
		entity.PaymentStatusPending, // retry
	},

	entity.PaymentStatusCompleted: {
		entity.PaymentStatusRefunded,
	},

	// final states: no outgoing transitions
	entity.PaymentStatusCanceled: {},
	entity.PaymentStatusExpired:  {},
	entity.PaymentStatusRefunded: {},
}

func IsValidTransition(current, next entity.PaymentStatus) bool {
	// idempotent update allowed
	if current == next {
		return true
	}

	allowed, ok := transitions[current]
	if !ok {
		return false
	}

	for _, v := range allowed {
		if v == next {
			return true
		}
	}

	return false
}

func IsFinal(status entity.PaymentStatus) bool {
	switch status {
	case entity.PaymentStatusCompleted,
		entity.PaymentStatusCanceled,
		entity.PaymentStatusExpired,
		entity.PaymentStatusRefunded:
		return true
	default:
		return false
	}
}
