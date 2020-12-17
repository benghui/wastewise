package main

import (
	"database/sql"
	"sync"
	"time"

	"github.com/gorilla/sessions"
)

// Server struct holding database connection pool
// and session cookiestore.
type Server struct {
	db    *sql.DB
	store *sessions.CookieStore
}

// Employees contain parameters for employee
type Employees struct {
	Mu        sync.RWMutex
	Username  string `json:"username"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Password  string `json:"password"`
	Role      string `json:"role"`
}

// Credentials contain parameters for logging in
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Products contain parameters for product
type Products struct {
	ProductID   int       `json:"product_id"`
	ProductName string    `json:"product_name"`
	CreatedDate time.Time `json:"created_date"`
	CostPrice   float64   `json:"cost_price"`
	SalesPrice  float64   `json:"sales_price"`
}

// ProductResult contains the parameters for products query
type ProductResult struct {
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
}

// WastageResult contains parameters for wastage combined with products & employees
type WastageResult struct {
	WastageID         int       `json:"wastage_id"`
	WastageDate       time.Time `json:"wastage_date"`
	WastageQuantity   int       `json:"quantity"`
	WastageReason     string    `json:"reason"`
	ProductName       string    `json:"product_name"`
	ProductCostPrice  float64   `json:"cost_price"`
	ProductSalesPrice float64   `json:"sales_price"`
	WastageLostSales  float64   `json:"lost_sales"`
	EmployeeFirstname string    `json:"firstname"`
}

// Wastage contains the parameters for wastage.
type Wastage struct {
	WastageID       int       `json:"wastage_id"`
	WastageDate     time.Time `json:"wastage_date"`
	WastageQuantity int       `json:"quantity"`
	WastageReason   string    `json:"reason"`
	ProductName     string    `json:"product_name"`
}

// WastageForm contains the parameters for creating & editing a wastage entry
type WastageForm struct {
	Mu              sync.RWMutex
	WastageDate     time.Time `json:"wastage_date"`
	WastageQuantity int       `json:"quantity"`
	WastageReason   string    `json:"reason"`
	ProductID       int       `json:"product_id"`
}

// ReportMonthly contains the parameters for monthly report
type ReportMonthly struct {
	Month          int     `json:"month"`
	WastageReason  string  `json:"reason"`
	ProductName    string  `json:"product_name"`
	TotalQuantity  int     `json:"total_quantity"`
	TotalLostSales float64 `json:"total_lost_sales"`
}
