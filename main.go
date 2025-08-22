package main

import (
	"errors"
	"fmt"
	"strings"
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
	MARKET OrderType = iota
	LIMIT
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
	Symbol            string
	Side              OrderSide
	Type              OrderType
	Quantity          int64
	RemainingQuantity int64
	Price             int64
	Status            OrderStatus
	CreateAt          time.Time
	UpdatedAt         time.Time
}

func parseStatus(s string) (OrderStatus, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "pending":
		return PENDING, nil
	case "partially_filled":
		return PARTIALLY_FILLED, nil
	case "filled":
		return FILLED, nil
	case "canceled":
		return CANCELLED, nil
	case "rejected":
		return REJECTED, nil
	}
	return 0, fmt.Errorf("invalid order status: %s", s)
}

func parseSide(s string) (OrderSide, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "buy":
		return BUY, nil
	case "sell":
		return SELL, nil
	}
	return 0, fmt.Errorf("invalid order side: %s", s)
}

func parseType(s string) (OrderType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "market":
		return MARKET, nil
	case "limit":
		return LIMIT, nil
	}
	return 0, fmt.Errorf("invalid order type: %s", s)
}

func validateQuantity(q int64) bool {
	return q > 0
}

func validatePrice(t OrderType, p int64) bool {
	if t == LIMIT && p <= 0 {
		return false
	}
	return true
}

func NewOrder(symbol, orderType, orderSide, orderStatus string, qty, price int64) (*Order, error) {
	status, err := parseStatus(orderStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	side, err := parseSide(orderSide)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	otype, err := parseType(orderType)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	if !validateQuantity(qty) {
		err := errors.New("qunatity must be greater than 0")
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	if !validatePrice(otype, price) {
		err := errors.New("price must be greater than 0")
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &Order{
		OrderID:           uuid.New(),
		Type:              otype,
		Symbol:            symbol,
		Side:              side,
		Status:            status,
		Quantity:          qty,
		RemainingQuantity: qty,
		Price:             price,
		CreateAt:          time.Now(),
		UpdatedAt:         time.Now(),
	}, nil
}

// Market orders must have Quantity > 0, Price can be ignored
// Limit orders must have both Quantity > 0 and Price > 0
// RemainingQuantity must always be <= Quantity
// Status transitions must follow valid state machine
// CreatedAt is immutable after creation
// Orders with RemainingQuantity = 0 should be marked FILLED

func main() {
	order, err := NewOrder("AAPL", "LIMIT", "BUY", "pending", 10, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v", order)
}
