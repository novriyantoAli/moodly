package task

const (
	// for generate bills monthly
	TypeGenerateMonthlyBills     = "bill:generate-monthly-bills"
	TypeGenerateBillPerSubscribe = "bill:generate-per-subscribe"

	// for check if unpaid bills now overdue
	TypeCheckUnpaidBills              = "bill:check-unpaid-bills"
	TypeChangeBillFromUnpaidToOverdue = "bill:change-unpaid-to-overdue"
)
