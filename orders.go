package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Order represents the model for an order
// Default table name will be `orders`
type Order struct {
	// gorm.Model
	OrderID      uint      `json:"orderId" gorm:"primary_key"`
	CustomerName string    `json:"customerName"`
	OrderedAt    time.Time `json:"orderedAt"`
	Items        []Item    `json:"items" gorm:"foreignkey:OrderID"`
}

// Item represents the model for an item in the order
type Item struct {
	// gorm.Model
	LineItemID  uint   `json:"lineItemId" gorm:"primary_key"`
	ItemCode    string `json:"itemCode"`
	Description string `json:"description"`
	Quantity    uint   `json:"quantity"`
	OrderID     uint   `json:"-"`
}

var db *gorm.DB

func initDB() {
	var err error
	dataSourceName := "root:redhat@tcp(localhost:3306)/?parseTime=True&autocommit=On"
	db, err = gorm.Open("mysql", dataSourceName)

	if err != nil {
		fmt.Println(err)
		panic("failed to connect database")
	}

	// Create the database. This is a one-time step.
	// Comment out if running multiple times - You may see an error otherwise
	db.Exec("CREATE DATABASE IF NOT EXISTS orders_db")
	db.Exec("USE orders_db")

	// Migration to create tables for Order and Item schema
	db.AutoMigrate(&Order{}, &Item{})
}

// func createOrder(w http.ResponseWriter, r *http.Request) {
func createOrder(c *gin.Context) {
	var order Order
	json.NewDecoder(c.Request.Body).Decode(&order)
	// Creates new order by inserting records in the `orders` and `items` table
	fmt.Println(1 << 32)
	order = Order{
		OrderID:      uint(rand.Int() % (1 << 32)), // get random id
		CustomerName: "Yangwawa",
		OrderedAt:    time.Now(),
		Items:        []Item{},
	}
	db.Save(&order)
	c.Header("Content-Type", "application/json")
	// json.NewEncoder(c.Writer).Encode(order)
	c.JSON(http.StatusOK, order)
}

func getOrders(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	var orders []Order
	db.Preload("Items").Find(&orders)
	c.JSON(http.StatusOK, orders)
}

func getOrder(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	inputOrderID := c.Params.ByName("orderId")
	var order Order
	db.Preload("Items").First(&order, inputOrderID)
	c.JSON(http.StatusOK, order)
}

func updateOrder(w http.ResponseWriter, r *http.Request) {
	var updatedOrder Order
	json.NewDecoder(r.Body).Decode(&updatedOrder)
	db.Save(&updatedOrder)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedOrder)
}

func deleteOrder(c *gin.Context) {
	inputOrderID := c.Params.ByName("orderId")
	// Convert `orderId` string param to uint64
	id64, _ := strconv.ParseUint(inputOrderID, 10, 64)
	// Convert uint64 to uint
	idToDelete := uint(id64)

	deleteOrderRows := db.Where("order_id = ?", idToDelete).Delete(&Item{}).RowsAffected
	deleteItemRows := db.Where("order_id = ?", idToDelete).Delete(&Order{}).RowsAffected
	if deleteItemRows >= 1 || deleteOrderRows >= 1 {
		c.JSON(http.StatusOK, map[string]string{
			"message": fmt.Sprintf("%d is successful deleted.", idToDelete),
		})
	} else {
		c.JSON(http.StatusAccepted, map[string]string{
			"message": fmt.Sprintf("%d is not exists in DB", idToDelete),
		})
	}
}

func main() {
	// router := mux.NewRouter()
	// // Create
	// router.HandleFunc("/orders", createOrder).Methods("POST")
	// // Read
	// router.HandleFunc("/orders/{orderId}", getOrder).Methods("GET")
	// // Read-all
	// router.HandleFunc("/orders", getOrders).Methods("GET")
	// // Update
	// router.HandleFunc("/orders/{orderId}", updateOrder).Methods("PUT")
	// // Delete
	// router.HandleFunc("/orders/{orderId}", deleteOrder).Methods("DELETE")
	router := gin.New()

	router.GET("/", createOrder)
	router.GET("/orders", getOrders)
	router.GET("/orders/:orderId", getOrder)
	router.DELETE("/orders/:orderId", deleteOrder)
	// Initialize db connection
	initDB()

	// log.Fatal(http.ListenAndServe(":8080", router))
	log.Fatal(router.Run())
}
