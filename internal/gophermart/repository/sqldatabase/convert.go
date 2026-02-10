package sqldatabase

import (
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
)

// ConvertModelOrderToDomainOrder performs conversion of [models.Order]
// to [domain.Order] object.
func ConvertModelOrderToDomainOrder(modelOrder models.Order) domain.Order {
	return domain.Order{
		ID:        domain.OrderID(modelOrder.ID),
		Status:    domain.OrderStatus(modelOrder.Status),
		CreatedAt: modelOrder.CreatedAt,
	}
}

func makeDomainOrders(modelOrders []models.Order) []domain.Order {
	orders := make([]domain.Order, 0, len(modelOrders))

	for _, modelOrder := range modelOrders {
		orders = append(orders, ConvertModelOrderToDomainOrder(modelOrder))
	}

	return orders
}

func makeDomainWithdrawalTransactions(modelWithdrawals []models.Withdrawal) []domain.WithdrawalTransaction {
	transactions := make([]domain.WithdrawalTransaction, 0, len(modelWithdrawals))

	for _, withdrawal := range modelWithdrawals {
		transactions = append(transactions, domain.WithdrawalTransaction{
			OrderID:     domain.OrderID(withdrawal.NextOrderID),
			Amount:      withdrawal.Amount,
			ProcessedAt: withdrawal.CreatedAt,
		})
	}

	return transactions
}
