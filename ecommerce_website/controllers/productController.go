package controllers

import (
	"net/http"

	"ecommerce/database"
	"ecommerce/model"

	"github.com/gin-gonic/gin"
)

func AddProduct(c *gin.Context) {
	// Parse the request body
	var newProduct model.CreateProduct
	if err := c.ShouldBindJSON(&newProduct); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the user
	product := model.Products{
		Category:    newProduct.Category,
		ProductName: newProduct.ProductName,
		Description: newProduct.Description,
		Price:       newProduct.Price,
		Attributes:  newProduct.Attributes,
		Inventory:   newProduct.Inventory,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed Add product"})
		return
	}

	// Return the token as a response
	c.JSON(http.StatusOK, gin.H{"data": product})
}

func GetProductByID(c *gin.Context) {
	id := c.Param("id")
	var product model.Products
	if err := database.DB.Where("id = ?", id).First(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Product_Details": product})
}

func GetProducts(c *gin.Context) {
	var products []model.Products
	database.DB.Find(&products)

	c.JSON(http.StatusOK, gin.H{"Products_Details": products})
}

func UpdateProductDetails(c *gin.Context) {
	//Get the user if exist
	id := c.Param("id")
	var product model.Products
	if err := database.DB.Where("id = ?", id).First(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	//validate input
	var new_details model.UpdateProducts
	if err := c.ShouldBindJSON(&new_details); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	product.Category = new_details.Category
	product.ProductName = new_details.ProductName
	product.Description = new_details.Description
	product.Price = new_details.Price
	product.Attributes = new_details.Attributes
	product.Inventory = new_details.Inventory
	database.DB.Save(&product)

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

func DeleteProduct(c *gin.Context) {

	id := c.Param("id")
	var product model.Products
	if err := database.DB.Where("id = ?", id).First(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	database.DB.Delete(&product)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func SearchProducts(c *gin.Context) {
	// Get the search query from the URL parameter
	query := c.Query("q")

	// Get the sort options from the URL parameters
	sortBy := c.DefaultQuery("sortBy", "addedDate")
	sortOrder := c.DefaultQuery("sortOrder", "asc")

	// Define a slice of valid sort options
	validSortOptions := []string{"addedDate", "price"}

	// Check if the sort option is valid
	validSortOption := false
	for _, option := range validSortOptions {
		if option == sortBy {
			validSortOption = true
			break
		}
	}
	if !validSortOption {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort option"})
		return
	}
	// Retrieve the products from the database
	var products []model.Products
	orderClause := ""
	if sortBy == "addedDate" {
		orderClause = "created_at"
	} else if sortBy == "price" {
		orderClause = "price"
	}

	if sortOrder == "asc" {
		orderClause += " ASC"
	} else if sortOrder == "desc" {
		orderClause += " DESC"
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sortOrder"})
		return
	}

	if orderClause == "" {
		// If the sort option is not recognized, default to sorting by addedDate in ascending order
		orderClause = "created_at ASC"
	}

	database.DB.Order(orderClause).Find(&products, "name LIKE ?", "%"+query+"%")

	// Return the sorted products
	c.JSON(http.StatusOK, gin.H{"products": products})
}
