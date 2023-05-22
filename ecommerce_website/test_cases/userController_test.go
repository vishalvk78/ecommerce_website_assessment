package controllers_test

import (
	"bytes"
	"ecommerce/controllers"
	"ecommerce/database"
	"ecommerce/model"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	// Set up the Gin router and create a test HTTP server
	router := gin.Default()
	router.POST("/login", controllers.Login)
	server := httptest.NewServer(router)
	defer server.Close()

	// Create a test user and add it to the database
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := model.Users{
		FullName: "Test User",
		Email:    "testuser@example.com",
		Password: string(hashedPassword),
		Role:     "user",
	}
	database.DB.Create(&user)
	defer database.DB.Delete(&user)

	// Define the test cases
	testCases := []struct {
		name            string
		requestBody     gin.H
		expectedStatus  int
		expectedMessage string
		expectedCookie  string
	}{
		{
			name: "Successful login",
			requestBody: gin.H{
				"email":    "testuser@example.com",
				"password": "password123",
			},
			expectedStatus:  http.StatusOK,
			expectedMessage: "success",
			expectedCookie:  "access_token",
		},
		{
			name: "Incorrect email",
			requestBody: gin.H{
				"email":    "wrongemail@example.com",
				"password": "password123",
			},
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "Invalid email or password",
			expectedCookie:  "",
		},
		{
			name: "Incorrect password",
			requestBody: gin.H{
				"email":    "testuser@example.com",
				"password": "wrongpassword",
			},
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "Invalid email or password",
			expectedCookie:  "",
		},
		{
			name: "Invalid request body",
			requestBody: gin.H{
				"invalid_field": "invalid_value",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Invalid request body",
			expectedCookie:  "",
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert the request body to JSON
			requestBodyBytes, _ := json.Marshal(tc.requestBody)

			// Send a request to the test server
			req, err := http.NewRequest(http.MethodPost, server.URL+"/login", bytes.NewReader(requestBodyBytes))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Check the response status code
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			// Check the response message
			var responseMap map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&responseMap)
			assert.Equal(t, tc.expectedMessage, responseMap["status"])

			// Check the cookies
			cookies := resp.Cookies()
			var cookieValue string
			for _, cookie := range cookies {
				if cookie.Name == tc.expectedCookie {
					cookieValue = cookie.Value
					break
				}
			}

			assert.Equal(t, tc.expectedCookie != "", len(cookies) > 0) // check if a cookie was set when expected
			assert.Equal(t, tc.expectedCookie, cookies[0].Name)        // check if the cookie name is correct
			assert.NotEmpty(t, cookieValue)                            // check if the cookie value is not empty

			// Send a request to the test server to get user information with the access token cookie
			req, err = http.NewRequest(http.MethodGet, server.URL+"/user", nil)
			assert.NoError(t, err)
			req.AddCookie(cookies[0])
			resp, err = http.DefaultClient.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Check the response status code
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// Check the response message
			var user model.Users
			err = json.NewDecoder(resp.Body).Decode(&user)
			assert.NoError(t, err)
			assert.Equal(t, "Test User", user.FullName)
			assert.Equal(t, "testuser@example.com", user.Email)
		})
	}
}
