package controllers

import (
	"net/http"

	"ecommerce/database"
	"ecommerce/model"
	"ecommerce/utils"

	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	// Parse the request body
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the user already exists
	var user model.Users
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err == nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Create the user
	user = model.Users{
		FullName: req.FullName,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     req.Role,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Return the token as a response
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func Login(ctx *gin.Context) {
	// Parse the request body to get the user credentials
	var user model.Users
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Check if the user with the provided email exists in the database
	var dbUser model.Users
	if err := database.DB.Where("email = ?", user.Email).First(&dbUser).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Compare the provided password with the hashed password in the database
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	config, _ := database.LoadConfig(".")

	// Generate Tokens
	expirationTime := time.Now().Add(24 * time.Hour) // Token expires in 24 hours
	access_token, err := utils.CreateToken(expirationTime, user.ID, config.AccessTokenPrivateKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	refresh_token, err := utils.CreateToken(expirationTime, user.ID, config.RefreshTokenPrivateKey)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	ctx.SetCookie("access_token", access_token, config.AccessTokenMaxAge*60, "/", "localhost", false, true)
	ctx.SetCookie("refresh_token", refresh_token, config.RefreshTokenMaxAge*60, "/", "localhost", false, true)
	ctx.SetCookie("logged_in", "true", config.AccessTokenMaxAge*60, "/", "localhost", false, false)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "access_token": access_token})
}

// Refresh Access Token
func RefreshAccessToken(ctx *gin.Context) {
	message := "could not refresh access token"

	cookie, err := ctx.Cookie("refresh_token")

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": message})
		return
	}

	config, _ := database.LoadConfig(".")

	sub, err := utils.ValidateToken(cookie, config.RefreshTokenPublicKey)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	var user model.Users
	result := database.DB.First(&user, "id = ?", fmt.Sprint(sub))
	if result.Error != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": "the user belonging to this token no logger exists"})
		return
	}

	expirationTime := time.Now().Add(time.Hour)
	access_token, err := utils.CreateToken(expirationTime, user.ID, config.AccessTokenPrivateKey)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	ctx.SetCookie("access_token", access_token, config.AccessTokenMaxAge*60, "/", "localhost", false, true)
	ctx.SetCookie("logged_in", "true", config.AccessTokenMaxAge*60, "/", "localhost", false, false)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "access_token": access_token})
}

func LogoutUser(ctx *gin.Context) {
	ctx.SetCookie("access_token", "", -1, "/", "localhost", false, true)
	ctx.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
	ctx.SetCookie("logged_in", "", -1, "/", "localhost", false, false)

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func GetUser(c *gin.Context) {
	id := c.Param("id")
	var user model.Users
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func GetUsers(c *gin.Context) {
	var users []model.Users
	database.DB.Find(&users)

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func UpdateUser(c *gin.Context) {
	//Get the user if exist
	id := c.Param("id")
	var user model.Users
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	}

	//validate input
	var input model.UpdateUser
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	user.FullName = input.FullName
	user.Email = input.Email
	user.Password = input.Password
	database.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func DeleteUser(c *gin.Context) {

	//get user if exist
	var user model.Users
	id := c.Param("id")
	if err := database.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	database.DB.Delete(&user)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func GetUserByID(userID int) (*model.Users, error) {
	user := &model.Users{}
	if err := database.DB.Where("id = ?", userID).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func GetWebTrackingCount1(c *gin.Context) {
	// Get the count of web tracking data
	var totalCount int64
	if err := database.DB.Model(&model.WebTracking{}).Count(&totalCount).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get web tracking count"})
		return
	}

	// Get the count of web tracking data for each product name
	rows, err := database.DB.Model(&model.WebTracking{}).Select("product, COUNT(*)").Group("product").Rows()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product counts"})
		return
	}
	defer rows.Close()

	// Build a map of product counts
	productCounts := make(map[string]int)
	for rows.Next() {
		var productName string
		var productCount int
		if err := rows.Scan(&productName, &productCount); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan product counts"})
			return
		}
		productCounts[productName] = productCount
	}

	// Return the product counts
	c.JSON(http.StatusOK, gin.H{
		"product_counts": productCounts,
		"total_count":    totalCount,
	})
}

func GetMe(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(model.Users)

	userResponse := &model.Users{
		ID:        currentUser.ID,
		FullName:  currentUser.FullName,
		Email:     currentUser.Email,
		Role:      currentUser.Role,
		CreatedAt: currentUser.CreatedAt,
		UpdatedAt: currentUser.UpdatedAt,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"user": userResponse}})
}

/*
	func GetWebTrackingCount1(c *gin.Context) {
		// Get the count of web tracking data
		var totalCount int64
		if err := database.DB.Model(&model.WebTracking{}).Count(&totalCount).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get web tracking count"})
			return
		}

		// Get the count of web tracking data for each product name
		var webTrackings []struct {
			ProductName string
			Count       int
		}

		if err := database.DB.Select("product_name, COUNT(*) as count").
			Group("product_name").
			Table("web_trackings").
			Scan(&webTrackings).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product counts"})
			return
		}

		// Build a map of product counts
		productCounts := make(map[string]int)
		for _, webTracking := range webTrackings {
			productCounts[webTracking.ProductName] = webTracking.Count
		}

		// Return the product counts
		c.JSON(http.StatusOK, gin.H{
			"Product_Name": productCounts,
		})
	}
*/
func GetWebTrackingCount(c *gin.Context) {
	// Get the count of web tracking data
	var totalCount int64
	if err := database.DB.Model(&model.WebTracking{}).Count(&totalCount).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get web tracking count"})
		return
	}

	// Get the count of web tracking data for each product name
	var webTrackings []struct {
		ProductID string
		Count     int
	}

	if err := database.DB.Select("product_id, COUNT(*) as count").
		Group("product_id").
		Table("web_trackings").
		Scan(&webTrackings).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product counts"})
		return
	}

	// Build a map of product counts
	productCounts := make(map[string]int)
	for _, webTracking := range webTrackings {
		productCounts[webTracking.ProductID] = webTracking.Count
	}

	// Return the product counts in the desired format
	response := make(map[string]int)
	for productID, count := range productCounts {
		if productID != "" {
			response[productID] = count
		}
	}

	c.JSON(http.StatusOK, gin.H{"ProductCounts": response})
}

func GetWebTrackig(c *gin.Context) {
	var web []model.WebTracking
	database.DB.Find(&web)

	c.JSON(http.StatusOK, gin.H{"Web": web})
}

func CreateWebTracking(c *gin.Context) {
	// Parse the request body
	var req model.Web
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the user
	track := model.WebTracking{
		ProductID: req.ProductID,
	}
	if err := database.DB.Create(&track).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Return the token as a response
	c.JSON(http.StatusOK, gin.H{"data": track})
}
