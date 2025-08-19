package orderbook

import (
	"time"

	"github.com/google/uuid"
)

type OrderSide int

const (
	BUY OrderSide = iota
	SELL
)

type OrderType int

const (
	Market OrderType = iota
	Limit
)

type OrderStatus int

const (
	PENDING OrderStatus = iota
	PARTIALLY_FILLED
	FILLED
	CANCELLED
	REJECTED
)

type Order struct {
	OrderID           uuid.UUID
	AccountID         uuid.UUID
	Symbol            string
	Side              OrderSide
	Type              OrderType
	Quantity          int64
	RemainingQUantity int64
	Price             int64
	Status            OrderStatus
	CreateAt          time.Time
	UpdatedAt         time.Time
}

// Market orders must have Quantity > 0, Price can be ignored
// Limit orders must have both Quantity > 0 and Price > 0
// RemainingQuantity must always be <= Quantity
// Status transitions must follow valid state machine
// CreatedAt is immutable after creation
// Orders with RemainingQuantity = 0 should be marked FILLED
