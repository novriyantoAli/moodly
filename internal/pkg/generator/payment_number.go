package generator

import (
	"crypto/rand"
	"fmt"
	"time"
)

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type PaymentNumberGenerator struct{}

func NewPaymentNumberGenerator() *PaymentNumberGenerator {
	return &PaymentNumberGenerator{}
}

func (g *PaymentNumberGenerator) Generate() string {
	b := make([]byte, 6)

	for i := range b {
		random := make([]byte, 1)
		_, _ = rand.Read(random)

		b[i] = chars[int(random[0])%len(chars)]
	}

	return fmt.Sprintf(
		"PAY-%s-%s",
		time.Now().Format("20060102"),
		string(b),
	)
}
