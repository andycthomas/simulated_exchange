package dto

// PlaceOrderRequest represents the request body for placing an order
type PlaceOrderRequest struct {
	Symbol   string  `json:"symbol" binding:"required,min=1,max=10"`
	Side     string  `json:"side" binding:"required,oneof=buy sell"`
	Type     string  `json:"type" binding:"required,oneof=limit market"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Price    float64 `json:"price" binding:"required,gt=0"`
}

// Validate performs additional business logic validation
func (r *PlaceOrderRequest) Validate() error {
	// Additional validation logic can be added here
	// For example, market orders might not require price
	return nil
}