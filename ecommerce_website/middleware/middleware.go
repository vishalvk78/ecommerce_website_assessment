package middleware

import (
	"ecommerce/controllers"
	"ecommerce/database"
	"ecommerce/model"
	"ecommerce/utils"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header from the request
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Authorization header is missing",
			})
			c.Abort()
			return
		}

		// Extract the JWT token from the Authorization header
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Parse the JWT token and extract the email and role claims
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("your-secret-key"), nil
		})

		// Get the current time
		currentTime := time.Now()

		// Extract the expiration time from the claims
		expValue, ok := claims["exp"]
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid JWT token: exp claim is missing",
			})
			c.Abort()
			return
		}

		// Convert the expiration time to a time.Time object
		var expirationTime time.Time
		switch expValue := expValue.(type) {
		case float64:
			expirationTime = time.Unix(int64(expValue), 0)
		case int64:
			expirationTime = time.Unix(expValue, 0)
		default:
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid JWT token: exp claim has an unsupported type",
			})
			c.Abort()
			return
		}
		// Compare the current time with the expiration time from the token
		if currentTime.After(expirationTime) {
			// Token has expired
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "JWT token has expired",
			})
			c.Abort()
			return
		}

		if err == nil || token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid JWT token",
			})
			c.Abort()
			return
		}

		email, ok := claims["email"].(string)
		if ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid JWT token",
			})
			c.Abort()
			return
		}
		/*
			roleClaim, ok := claims["role"].(string)
			if !ok || roleClaim != role {
				c.JSON(http.StatusUnauthorized, gin.H{
					"message": "Unauthorized access",
				})
				c.Abort()
				return
			}
		*/
		// Add the email to the context for use in the handler
		c.Set("email", email)

		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the user token from the request header
		userToken := c.Request.Header.Get("Authorization")
		log.Println("User token:", userToken)

		// Identify the user based on the user token
		var userID int
		if userToken == "" {
			// If there is no user token, assume it's a guest user
			userID = 0
		} else {
			// If there is a user token, extract the user ID from it
			// and use it to identify the user
			claims, err := utils.ParseToken(userToken)
			if err != nil {
				//log.Println("Error parsing token:", err)
				//c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user token"})
				return

			}

			userID = claims

		}

		// Check if user is an admin
		user, err := controllers.GetUserByID(userID)
		if err != nil {
			log.Println("Error getting user:", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
			return
		}
		if user.Role != "admin" {
			log.Println("User is not an admin")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User is not an admin"})
			return
		}

		// If user is an admin, continue with the request
		c.Next()
	}
}

func WebTrackingMiddleware(productField string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the product ID from the request
		productID := c.Param(productField)

		if productID == "" {
			fmt.Println("productID is empty")
		}

		product, err := GetProductsByID(productID)
		if err != nil {
			//c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		// Capture the user behavior data
		webTrackingData := model.WebTracking{
			ProductID:   productID,
			Category:    product.Category,
			ProductName: product.ProductName,
			Attributes:  product.Attributes,
		}

		// Save the user behavior data to the database or log file
		if err := database.DB.Create(&webTrackingData).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user behavior data"})
			return
		}

		// Continue with the request
		c.Next()

		// Update the product counts
		if productID != "" {
			productCounts, err := getProductCounts()
			if err != nil {
				log.Println("Failed to get product counts:", err)
			} else {
				productCounts[productID]++
				//setProductCounts(productCounts)
			}
		}
	}
}

func getProductCounts() (map[string]int, error) {
	// Get the count of web tracking data for each product name
	var webTrackings []struct {
		ProductID string
		Count     int
	}
	if err := database.DB.Table("web_trackings").
		Select("product_id, COUNT(*) as count").
		Group("product_id").
		Scan(&webTrackings).Error; err != nil {
		return nil, err
	}

	// Build a map of product counts
	productCounts := make(map[string]int)
	for _, webTracking := range webTrackings {
		if webTracking.ProductID != "" {
			productCounts[webTracking.ProductID] = webTracking.Count
		}
	}

	// Remove the empty key entry if present
	delete(productCounts, "")

	return productCounts, nil
}

/*
func setProductCounts(productCounts map[string]int) {
	// Update the count of web tracking data for each product name
	for productID, count := range productCounts {
		if err := database.DB.Model(&model.Products{}).Where("name = ?", productID).UpdateColumn("count", count).Error; err != nil {
			log.Println("Failed to update product count:", err)
		}
	}
}
*/

func GetProductsByID(productID string) (model.Products, error) {
	// Check if productID is empty
	if productID == "" {
		return model.Products{}, errors.New("productID is empty")
	}

	// Query the database for the product with the given ID
	var product model.Products
	if err := database.DB.Where("id = ?", productID).First(&product).Error; err != nil {
		return product, err
	}
	return product, nil
}
