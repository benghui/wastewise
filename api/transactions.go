package main

import (
	"database/sql"
	"log"
	"time"
)

// QueryPassword returns the pointer to hashed password in database for a given username.
func QueryPassword(db *sql.DB, username string) (*string, *int, *string, error) {
	// Declare password variable as pointer to string
	var (
		password   *string
		employeeID *int
		role       *string
	)
	row := db.QueryRow("SELECT password, employee_id, role FROM employees WHERE username=?", username)
	err := row.Scan(&password, &employeeID, &role)

	switch err {
	case sql.ErrNoRows:
		log.Println("No rows returned!")
		return nil, nil, nil, err
	case nil:
		return password, employeeID, role, nil
	default:
		return nil, nil, nil, err
	}
}

// QueryWastage returns the pointer to a slice of WastageResult.
func QueryWastage(db *sql.DB) ([]*WastageResult, error) {
	queryString := `SELECT wastage.wastage_id, wastage.wastage_date, wastage.quantity, wastage.reason,
		products.product_name, products.cost_price, products.sales_price,
		(products.sales_price * wastage.quantity) AS lost_sales, employees.firstname
		FROM wastage, products, employees
		WHERE wastage.product_id=products.product_id
		AND wastage.employee_id=employees.employee_id
		AND wastage.wastage_date BETWEEN NOW() - INTERVAL 7 DAY AND NOW()
		ORDER BY wastage.wastage_id DESC`

	rows, err := db.Query(queryString)
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
func CreateWastageEntry(db *sql.DB, newWastageDate time.Time, newWastageQuantity int, newWastageReason string, productID int, employeeID int) error {
	stmt, err := db.Prepare(`INSERT INTO wastage (wastage_date, quantity, reason, product_id, employee_id) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(newWastageDate, newWastageQuantity, newWastageReason, productID, employeeID)
	if err != nil {
		return err
	}
	return nil
}

// QuerySingleWastage returns a pointer to Wastage.
func QuerySingleWastage(db *sql.DB, id int) (*Wastage, error) {
	queryString := `SELECT wastage.wastage_id, wastage.wastage_date, wastage.quantity, wastage.reason, products.product_name
	FROM wastage, products
	WHERE wastage.wastage_id=?
	AND wastage.product_id=products.product_id`
	row := db.QueryRow(queryString, id)

	wastage := &Wastage{}

	err := row.Scan(
		&wastage.WastageID,
		&wastage.WastageDate,
		&wastage.WastageQuantity,
		&wastage.WastageReason,
		&wastage.ProductName)

	switch err {
	case sql.ErrNoRows:
		return nil, err
	case nil:
		return wastage, nil
	default:
		return nil, err
	}
}

// UpdateWastageEntry updates an existing wastage entry.
func UpdateWastageEntry(db *sql.DB, id int, editWastageDate time.Time, editWastageQuantity int, editWastageReason string, productID int, employeeID int) error {
	stmt, err := db.Prepare("UPDATE wastage SET wastage_date=?, quantity=?, reason=?, product_id=?, employee_id=? WHERE wastage_id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(editWastageDate, editWastageQuantity, editWastageReason, productID, employeeID, id)
	if err != nil {
		return err
	}
	return nil
}

// QueryWastageReportMonthly returns a pointer to ReportMonthly
func QueryWastageReportMonthly(db *sql.DB) ([]*ReportMonthly, error) {
	queryString := `SELECT MONTH(wastage.wastage_date) AS month, products.product_name,
	SUM(wastage.quantity) AS total_quantity, SUM(wastage.quantity * products.sales_price) AS  total_lost_sales
	FROM wastage, products
	WHERE wastage.product_id=products.product_id
	GROUP BY products.product_name, MONTH(wastage.wastage_date)
	ORDER BY total_lost_sales DESC;`

	rows, err := db.Query(queryString)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	reports := []*ReportMonthly{}

	for rows.Next() {
		result := &ReportMonthly{}

		err := rows.Scan(
			&result.Month,
			&result.ProductName,
			&result.TotalQuantity,
			&result.TotalLostSales)
		if err != nil {
			return nil, err
		}

		reports = append(reports, result)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}
