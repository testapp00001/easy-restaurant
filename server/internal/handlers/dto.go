package handlers

// NewCookingItemDTO is the data structure sent to the kitchen via WebSocket
// when a new item is approved for cooking.
type NewCookingItemDTO struct {
	OrderItemID      uint   `json:"order_item_id"`
	DishName         string `json:"dish_name"`
	Quantity         int    `json:"quantity"`
	Notes            string `json:"notes"`
	TableNumber      string `json:"table_number"`
	KitchenGroupName string `json:"kitchen_group_name"`
}
