package main

import "time"

// Employees contain parameters for employee
type Employees struct {
	EmployeeID int    `json:"employee_id"`
	Username   string `json:"username"`
	Firstname  string `json:"firstname"`
	Lastname   string `json:"lastname"`
	Password   string `json:"password"`
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
