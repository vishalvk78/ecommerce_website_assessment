package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce/database"
	"ecommerce/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAddProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	// Create a config object for the database connection
	config := &database.Config{
		DBHost:         "localhost",
		DBUserName:     "root",
		DBUserPassword: "mysql",
		DBName:         "ecommerce_test",
		DBPort:         "5432",
	}

	// Connect to the test database
	database.ConnectDB(config)

	t.Run("should add a new product", func(t *testing.T) {
		// Define the request body
		newProduct := model.CreateProduct{
			Category:    "Books",
			ProductName: "The Art of Computer Programming",
			Description: "A comprehensive monograph written by Donald Knuth",
			Price:       50.0,
		}
		body, _ := json.Marshal(newProduct)

		// Perform the request
		req, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check the response
		assert.Equal(t, http.StatusOK, w.Code)
		var response struct {
			Data model.Products `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, newProduct.Category, response.Data.Category)
		assert.Equal(t, newProduct.ProductName, response.Data.ProductName)
		assert.Equal(t, newProduct.Description, response.Data.Description)
		assert.Equal(t, newProduct.Price, response.Data.Price)
	})

	t.Run("should return an error with invalid request body", func(t *testing.T) {
		// Define the request body
		invalidBody := []byte(`{"category": "Books"}`)

		// Perform the request
		req, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewReader(invalidBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check the response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response struct {
			Error string `json:"error"`
		}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Key: 'CreateProduct.ProductName' Error:Field validation for 'ProductName' failed on the 'required' tag", response.Error)
	})
}

func TestGetProductByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	// Create a config object for the database connection
	config := &database.Config{
		DBHost:         "localhost",
		DBUserName:     "root",
		DBUserPassword: "mysql",
		DBName:         "ecommerce_test",
		DBPort:         "5432",
	}

	// Connect to the test database
	database.ConnectDB(config)

	// Add a new product to the database
	newProduct := model.CreateProduct{
		Category:    "Books",
		ProductName: "The Art of Computer Programming",
		Description: "A comprehensive monograph written by Donald Knuth",
		Price:       50.0,
	}
	database.DB.Create(&model.Products{
		Category:    newProduct.Category,
		ProductName: newProduct.ProductName,
		Description: newProduct.Description,
		Price:       newProduct.Price,
	})

	t.Run("should return a product with valid ID", func(t *testing.T) {
		// Perform the request
		req, _ := http.NewRequest(http.MethodGet, "/products/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check the response
		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			ProductDetails model.Products `json:"Product_Details"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Error unmarshaling response body: %s", err.Error())
		}

		assert.Equal(t, "The Art of Computer Programming", response.ProductDetails.ProductName)
		assert.Equal(t, "A comprehensive monograph written by Donald Knuth", response.ProductDetails.Description)
		assert.Equal(t, 50.0, response.ProductDetails.Price)
	})

	t.Run("should return error for invalid ID", func(t *testing.T) {
		// Perform the request
		req, _ := http.NewRequest(http.MethodGet, "/products/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check the response
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Error unmarshaling response body: %s", err.Error())
		}

		assert.Equal(t, "Record not found!", response["error"])
	})
}
