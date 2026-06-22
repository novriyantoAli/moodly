package contract

type PaymentNumberGenerator interface {
	Generate() string
}
