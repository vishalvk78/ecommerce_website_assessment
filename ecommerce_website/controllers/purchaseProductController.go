package controllers

import (
	"ecommerce/database"
	"ecommerce/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PurchaseProduct handles the purchase of a product
func PurchaseProduct(ctx *gin.Context) {
	// Retrieve product information from the database based on the product ID
	productID := ctx.Param("id")

	product, err := GetProductsByIDs(productID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Perform the purchase operation
	err = PerformPurchase(product)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error purchasing product"})
		return
	}

	// Return a success response to the user
	ctx.JSON(http.StatusOK, gin.H{"message": "Product purchased successfully"})
}

// PerformPurchase performs the purchase operation for a product
func PerformPurchase(product model.Products) error {
	// Update the product inventory
	err := UpdateProductInventory(product)
	if err != nil {
		return err
	}

	// Create an order
	order := model.Order{
		ProductID:  product.ID,
		Quantity:   1,
		TotalPrice: product.Price,
		// Set other order details (e.g., user ID, shipping address, etc.)
	}
	err = CreateOrder(order)
	if err != nil {
		// In case of error, you may need to handle the inventory rollback or compensate for the failed purchase
		return err
	}

	return nil
}

// UpdateProductInventory updates the inventory of a product
func UpdateProductInventory(product model.Products) error {
	// Decrement the product inventory by 1
	product.Inventory--

	// Update the product inventory in the database
	if err := database.DB.Model(&product).Update("inventory", product.Inventory).Error; err != nil {
		return err
	}

	return nil
}

// CreateOrder creates a new order
func CreateOrder(order model.Order) error {
	// Create the order in the database
	if err := database.DB.Create(&order).Error; err != nil {
		return err
	}

	return nil
}

func GetProductsByIDs(productID string) (model.Products, error) {
	// Query the database for the product with the given ID
	var product model.Products
	if err := database.DB.Where("id = ?", productID).First(&product).Error; err != nil {
		return product, err
	}
	return product, nil
}

func GetAllOrders(c *gin.Context) {
	var order []model.Order
	database.DB.Find(&order)

	c.JSON(http.StatusOK, gin.H{"Products_Details": order})
}
