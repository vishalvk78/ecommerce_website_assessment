package main

import (
	"ecommerce/database"
	"ecommerce/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	config, err := database.LoadConfig(".")
	if err != nil {
		log.Fatal("🚀 Could not load environment variables", err)
	}

	//Initilize the database connection
	database.ConnectDB(&config)

	//defer database.Close()

	//Initilize the router
	router := gin.Default()

	//setup routes
	//api := .Group("/api")

	routes.ProductRoutes(router)
	routes.UserRoutes(router)

	//start the servers
	err = router.Run(":8080")
	if err != nil {
		log.Fatal("Failed to start the server:", err)
	}

}
