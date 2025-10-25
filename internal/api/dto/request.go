package dto

import "fmt"

// PlaceOrderRequest represents the request body for placing an order
type PlaceOrderRequest struct {
	Symbol   string  `json:"symbol" binding:"required,min=1,max=10"`
	Side     string  `json:"side" binding:"required,oneof=buy sell"`
	Type     string  `json:"type" binding:"required,oneof=limit market"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Price    float64 `json:"price"` // No binding validation - validated in Validate()
}

// Validate performs additional business logic validation
func (r *PlaceOrderRequest) Validate() error {
	// Limit orders must have a price greater than 0
	if r.Type == "limit" && r.Price <= 0 {
		return fmt.Errorf("limit orders require a price greater than 0")
	}

	// Market orders can have any price (it will be ignored by the engine)
	return nil
}
