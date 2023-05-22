package controllers

import (
	"ecommerce/database"
	"ecommerce/model"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetProductDetails(ctx *gin.Context) {
	// Retrieve product information from the database based on the product ID
	productID := ctx.Param("id")
	product, err := GetProductsByID(productID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Retrieve related products based on the recommendation algorithm
	relatedProducts, err := GetRelatedProducts(product)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving related products"})
		return
	}

	// Display the product details and recommended products to the user
	ctx.JSON(http.StatusOK, gin.H{
		"product":           product,
		"recommended_items": relatedProducts,
	})
}

func GetRelatedProducts(product model.Products) ([]model.Products, error) {
	// Retrieve all products with the same category as the input product
	relatedProducts, err := GetProductsByCategory(product.Category)
	if err != nil {
		return nil, err
	}

	// Check if relatedProducts is empty
	if len(relatedProducts) == 0 {
		return nil, errors.New("no related products found")
	}

	// Collect user behavior data
	userBehaviorData, err := GetUserBehaviorData()
	if err != nil {
		return nil, err
	}

	// Group behaviors by product ID
	behaviorCounts := make(map[string]int)
	for _, behavior := range userBehaviorData {
		// Only consider behaviors related to the same category as the input product
		if behavior.Category == product.Category {
			behaviorCounts[behavior.ProductID]++
		}
	}

	// Sort behaviorCounts by occurrence count in descending order
	sortedBehaviors := make([]struct {
		ProductID string
		Count     int
	}, 0, len(behaviorCounts))
	for productID, count := range behaviorCounts {
		sortedBehaviors = append(sortedBehaviors, struct {
			ProductID string
			Count     int
		}{ProductID: productID, Count: count})
	}
	sort.Slice(sortedBehaviors, func(i, j int) bool {
		return sortedBehaviors[i].Count > sortedBehaviors[j].Count
	})

	// Retrieve the top 5 related products by ID (excluding the input product)
	var relatedProductIDs []string
	for i := 0; i < len(sortedBehaviors) && len(relatedProductIDs) < 5; i++ {
		if sortedBehaviors[i].ProductID != strconv.Itoa(product.ID) {
			relatedProductIDs = append(relatedProductIDs, sortedBehaviors[i].ProductID)
		}
	}

	// Retrieve the related products by ID
	relatedProducts, err = GetProductsByIDs1(relatedProductIDs)
	if err != nil {
		return nil, err
	}

	return relatedProducts, nil
}

func GetProductsByCategory(category string) ([]model.Products, error) {
	// Query the database for all products with the given category
	// and return the results
	var products []model.Products
	if err := database.DB.Where("category = ?", category).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func GetProductsByID(productID string) (model.Products, error) {
	// Query the database for the product with the given ID
	var product model.Products
	if err := database.DB.Where("id = ?", productID).First(&product).Error; err != nil {
		return product, err
	}
	return product, nil
}

func GetProductsByIDs1(productIDs []string) ([]model.Products, error) {
	// Query the database for all products with the given IDs
	// and return the results
	var products []model.Products
	if err := database.DB.Where("id IN (?)", productIDs).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func GetUserBehaviorData() ([]model.WebTracking, error) {
	// Query the database for all user behavior data
	var userBehaviorData []model.WebTracking
	if err := database.DB.Find(&userBehaviorData).Error; err != nil {
		return nil, err
	}
	return userBehaviorData, nil
}
