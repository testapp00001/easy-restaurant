package handlers

import (
	"log"
	"restaurant-api/internal/database"
	"restaurant-api/internal/hub"
	"restaurant-api/internal/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

// GetPendingOrders retrieves all orders awaiting manager approval
func GetPendingOrders(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var orders []models.Order
		// Preload related data for the kitchen view
		db.Preload("OrderItems.MenuItem").Preload("RestaurantTable").
			Where("status = ?", "PendingApproval").Find(&orders)
		return c.JSON(orders)
	}
}

type ReviewOrderRequest struct {
	Items []struct {
		OrderItemID uint   `json:"order_item_id"`
		Action      string `json:"action"` // "Approve" or "Reject"
		Reason      string `json:"reason,omitempty"`
	} `json:"items"`
}

// ReviewOrder allows a manager to approve or reject items in an order
func ReviewOrder(db *gorm.DB, h *hub.Hub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req ReviewOrderRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		var orderID uint
		var order models.Order
		var customerSessionID uint

		// Use a transaction to ensure all updates are atomic
		err := db.Transaction(func(tx *gorm.DB) error {
			for _, item := range req.Items {
				var orderItem models.OrderItem
				// Find the order item and its related order and session
				if err := tx.Preload("Order.UserSession").First(&orderItem, item.OrderItemID).Error; err != nil {
					return fiber.NewError(fiber.StatusNotFound, "Order item not found")
				}

				// Set order details from the first item processed
				if orderID == 0 {
					orderID = orderItem.OrderID
					order = orderItem.Order
					customerSessionID = orderItem.Order.UserSessionID
				}

				switch item.Action {
				case "Approve":
					orderItem.Status = "Approved"
				case "Reject":
					orderItem.Status = "Rejected"
					orderItem.RejectedReason = item.Reason
				}

				if err := tx.Save(&orderItem).Error; err != nil {
					return err // Rollback
				}
			}
			// Update the main order status after processing all items
			return tx.Model(&order).Update("status", "Approved").Error
		})

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// After transaction is successful, broadcast WebSocket events
		go notifyKitchenAndCustomer(orderID, customerSessionID, h)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Order reviewed"})
	}
}

// notifyKitchenAndCustomer fetches the final state of the order and notifies relevant parties
func notifyKitchenAndCustomer(orderID, sessionID uint, h *hub.Hub) {
	var reviewedOrderItems []models.OrderItem
	database.DB.Debug().Preload("MenuItem.Category").Where("order_id = ?", orderID).Find(&reviewedOrderItems)

	for _, item := range reviewedOrderItems {
		switch item.Status {
		case "Approved":
			// Notify the specific kitchen group
			// Create the lean DTO for the kitchen
			cookingItem := NewCookingItemDTO{
				OrderItemID:      item.ID,
				DishName:         item.MenuItem.Name,
				Quantity:         item.Quantity,
				Notes:            item.CustomizationNotes,
				TableNumber:      item.Order.RestaurantTable.TableNumber,   // Get table number from preloaded data
				KitchenGroupName: item.MenuItem.Category.KitchenGroup.Name, // Get group name from preloaded data
			}

			// Notify the specific kitchen group with the DTO
			room := "kitchen_group_" + strconv.Itoa(int(item.MenuItem.Category.KitchenGroupID))
			h.BroadcastToRoom(room, "new_item_for_cooking", cookingItem)
		case "Rejected":
			// Notify the customer
			// The DTO for customer notification can also be improved
			rejectionInfo := fiber.Map{
				"dish_name": item.MenuItem.Name,
				"reason":    item.RejectedReason,
			}

			// Notify the customer
			room := "customer_session_" + strconv.Itoa(int(sessionID))
			h.BroadcastToRoom(room, "item_rejected", rejectionInfo)
		}
	}
}

// *Note: The `notifyKitchenAndCustomer` function is placed outside the handler to be run as a goroutine,
// ensuring the API responds quickly without waiting for WebSocket broadcasts to complete.

// GetGroupItems retrieves all approved items for a specific kitchen group
func GetGroupItems(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		groupID := c.Params("id")

		var items []models.OrderItem
		// This is a more complex query joining through MenuItem and Category
		db.Joins("MenuItem").Joins("MenuItem.Category").
			Where("order_items.status = ? AND \"MenuItem__Category\".kitchen_group_id = ?", "Approved", groupID).
			Find(&items)

		return c.JSON(items)
	}
}

// AssignOrderItem allows a chef to claim an item
func AssignOrderItem(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orderItemID, _ := strconv.Atoi(c.Params("id"))
		chefID := c.Locals("user_id").(uint) // Assuming user_id is uint from middleware

		result := db.Model(&models.OrderItem{}).
			Where("id = ? AND status = ?", orderItemID, "Approved").
			Updates(map[string]interface{}{"status": "Assigned", "assigned_chef_id": chefID})

		if result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to assign item"})
		}
		if result.RowsAffected == 0 {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Item could not be assigned (may already be assigned or not approved)"})
		}

		// In a real app, you would also broadcast this status update to the manager

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
	}
}

// UpdateOrderItemStatus allows a chef to update an item's cooking status
func UpdateOrderItemStatus(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orderItemID, _ := strconv.Atoi(c.Params("id"))
		chefID := c.Locals("user_id").(uint)

		var req struct {
			Status string `json:"status"` // "Cooking" or "Ready"
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		result := db.Model(&models.OrderItem{}).
			Where("id = ? AND assigned_chef_id = ?", orderItemID, chefID).
			Update("status", req.Status)

		if result.Error != nil || result.RowsAffected == 0 {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Failed to update status"})
		}

		// Broadcast status update to manager
		// hub.WSHub.BroadcastToRoom("kitchen_manager", "item_status_update", ...)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
	}
}

// WebSocketHandler now accepts the hub as a dependency and returns the actual handler
func WebSocketHandler(h *hub.Hub) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		// For now, we will assume the room is passed as a query param for simplicity.
		// In a real app, you would verify a JWT token here to get the user's role and ID.
		room := c.Query("room")
		if room == "" {
			log.Println("Room query param is required")
			c.Close()
			return
		}

		client := &hub.Client{Conn: c, Room: room}
		h.Register(client) // Use the passed-in hub 'h'

		defer func() {
			h.Unregister(client) // Use the passed-in hub 'h'
			c.Close()
		}()

		for {
			// Read message from client (can be used for pings or other interactions)
			_, _, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("read error:", err)
				}
				break // Exit the loop on error
			}
		}
	}
}
