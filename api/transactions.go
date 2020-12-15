package main

import (
	"database/sql"
	"log"
)

// QueryPassword returns the pointer to hashed password in database for a given username.
func QueryPassword(db *sql.DB, username string) (*string, *int, error) {
	// Declare password variable as pointer to string
	var (
		password   *string
		employeeID *int
	)
	row := db.QueryRow("SELECT password, employee_id FROM employees WHERE username=?", username)
	err := row.Scan(&password, &employeeID)

	switch err {
	case sql.ErrNoRows:
		log.Println("No rows returned!")
		return nil, nil, err
	case nil:
		return password, employeeID, nil
	default:
		return nil, nil, err
	}
}

// QueryWastage returns the pointer to a slice of WastageResult.
func QueryWastage(db *sql.DB) ([]*WastageResult, error) {
	rows, err := db.Query(
		`SELECT wastage.wastage_id, wastage.wastage_date, wastage.quantity, wastage.reason,
		products.product_name, products.cost_price, products.sales_price,
		(products.sales_price * wastage.quantity) AS lost_sales, employees.firstname
		FROM wastage, products, employees
		WHERE wastage.product_id=products.product_id
		AND wastage.employee_id=employees.employee_id
		AND wastage.wastage_date BETWEEN NOW() - INTERVAL 7 DAY AND NOW()
		ORDER BY wastage.wastage_id DESC`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	wastageResults := []*WastageResult{}

	for rows.Next() {
		result := &WastageResult{}

		err := rows.Scan(
			&result.WastageID,
			&result.WastageDate,
			&result.WastageQuantity,
			&result.WastageReason,
			&result.ProductName,
			&result.ProductCostPrice,
			&result.ProductSalesPrice,
			&result.WastageLostSales,
			&result.EmployeeFirstname)
		if err != nil {
			return nil, err
		}

		wastageResults = append(wastageResults, result)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return wastageResults, nil
}

// QueryProducts returns the pointer to a slice of ProductResult.
func QueryProducts(db *sql.DB) ([]*ProductResult, error) {
	rows, err := db.Query(`SELECT product_id, product_name FROM products`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productResults := []*ProductResult{}

	for rows.Next() {
		result := &ProductResult{}

		err := rows.Scan(&result.ProductID, &result.ProductName)
		if err != nil {
			return nil, err
		}

		productResults = append(productResults, result)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return productResults, nil
}

// CreateWastageEntry writes to database a new wastage entry.
func CreateWastageEntry(db *sql.DB, newWastageQuantity int, newWastageReason string, productID int, employeeID int) (error) {
	stmt, err := db.Prepare(`INSERT INTO wastage (wastage_date, quantity, reason, product_id, employee_id) VALUES (NOW(), ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(newWastageQuantity, newWastageReason, productID, employeeID)
	if err != nil {
		return err
	}
	return nil
}