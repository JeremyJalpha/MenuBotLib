package menubotlib

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type CustomerOrder struct {
	OrderID           int
	CellNumber        string
	CatalogueID       string
	OrderItems        OrderItems
	OrderTotal        int
	IsPaid            bool
	DateTimeDelivered sql.NullTime
	IsClosed          bool
}

var ErrNoRows = errors.New("no rows found")

func (c *CustomerOrder) SetCurrentOrderFromDB(db *sql.DB, senderNum string, isAutoInc bool) error {
	var orderItemsJSON []byte

	c.CellNumber = senderNum
	queryString := `SELECT orderid, cellnumber, catalogueID, orderitems, ispaid, datetimedelivered 
                    FROM CustomerOrder 
                    WHERE cellnumber = $1 AND isclosed = false
                    ORDER BY orderid DESC
                    LIMIT 1`
	row := db.QueryRow(queryString, c.CellNumber)
	err := row.Scan(&c.OrderID, &c.CellNumber, &c.CatalogueID, &orderItemsJSON, &c.IsPaid, &c.DateTimeDelivered)
	if err != nil {
		if err == sql.ErrNoRows {
			if !isAutoInc {
				// No rows were found, get the next value in the sequence
				err = db.QueryRow("SELECT nextval('customerorder_id_seq')").Scan(&c.OrderID)
				if err != nil {
					return err
				}
				return ErrNoRows
			}
			return ErrNoRows
		} else {
			// Some other error occurred
			return err
		}
	} else {
		// Unmarshal JSON data into the OrderItems struct
		err = json.Unmarshal(orderItemsJSON, &c.OrderItems)
		if err != nil {
			return fmt.Errorf("failed to unmarshal orderItems: %w", err)
		}
	}

	return nil
}

func (c *CustomerOrder) checkInitialization(db *sql.DB, senderNum string, isAutoInc bool) string {
	//Get the customer's current order
	if c.OrderItems.MenuIndications == nil {
		c.SetCurrentOrderFromDB(db, senderNum, isAutoInc)
		if c.OrderItems.MenuIndications == nil {
			return "No current order. We vill asks ze questions."
		}
	}
	return custOrderInitState
}

// A function that returns the current order of a user as a string
func (c *CustomerOrder) GetCurrentOrderAsAString(db *sql.DB, senderNum string, isAutoInc bool) string {
	isInited := c.checkInitialization(db, senderNum, isAutoInc)
	if isInited != custOrderInitState {
		return isInited
	}
	// iterate over the slice and build the string
	var orderItemsString string
	for _, item := range c.OrderItems.MenuIndications {
		orderItemsString += fmt.Sprintf("\n%d: %s,", item.ItemMenuNum, item.ItemAmount)
	}
	// If DateTimeDelivered is null then set it to "Not yet delivered"
	var dateTimeDelivered string
	if c.DateTimeDelivered.Valid {
		dateTimeDelivered = c.DateTimeDelivered.Time.Format("2006-01-02 15:04:05")
	} else {
		dateTimeDelivered = "Not yet delivered"
	}
	return fmt.Sprintf("Is Paid: %t\nDelivered on: %v\nOrder Items:%s",
		c.IsPaid, dateTimeDelivered, orderItemsString)
}

// Insert User Answer into database
func (c *CustomerOrder) insertOrder(db *sql.DB) error {
	// Convert OrderItems struct to JSON string
	orderItemsJSON, err := json.Marshal(c.OrderItems)
	if err != nil {
		return fmt.Errorf("failed to marshal orderItems: %w", err)
	}

	// Prepare an SQL statement to insert a new order
	queryString := `INSERT INTO CustomerOrder (orderid, cellnumber, catalogueID, orderitems, ispaid, datetimedelivered, isclosed) 
                    VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = db.Exec(queryString, c.OrderID, c.CellNumber, c.CatalogueID, orderItemsJSON, c.IsPaid, c.DateTimeDelivered, c.IsClosed)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	return nil
}

func (c *CustomerOrder) updateCurrentOrder(db *sql.DB) error {
	// Convert OrderItems struct to JSON string
	orderItemsJSON, err := json.Marshal(c.OrderItems)
	if err != nil {
		return fmt.Errorf("failed to marshal orderItems: %w", err)
	}

	// Prepare an SQL statement to update the order
	queryString := `UPDATE CustomerOrder SET cellnumber = $1, catalogueID = $2, orderitems = $3, ispaid = $4, datetimedelivered = $5, isclosed = $6 WHERE orderid = $7`
	_, err = db.Exec(queryString, c.CellNumber, c.CatalogueID, orderItemsJSON, c.IsPaid, c.DateTimeDelivered, c.IsClosed, c.OrderID)
	if err != nil {
		return err
	}

	return nil
}

// UpdateOrInsertCurrentOrder updates or inserts a customer order in the database.
func (c *CustomerOrder) cleanOrderItems() error {
	// Filter out items with ItemAmount equal to "0"
	var filteredMenu []MenuIndication
	for _, ordItm := range c.OrderItems.MenuIndications {
		if ordItm.ItemAmount != "0" {
			filteredMenu = append(filteredMenu, ordItm)
		}
	}

	// Update c.OrderItems.MenuIndications with the filtered slice
	c.OrderItems.MenuIndications = filteredMenu

	return nil
}

// UpdateOrInsertCurrentOrder updates or inserts a customer order in the database.
func (c *CustomerOrder) UpdateCustOrdItems(update OrderItems) error {
	// Initialize a map to track processed ItemMenuNum values
	processed := make(map[int]bool)

	for _, upd := range update.MenuIndications {
		for i, ordItm := range c.OrderItems.MenuIndications {
			if ordItm.ItemMenuNum == upd.ItemMenuNum {
				c.OrderItems.MenuIndications[i] = upd // Overwrite existing ordItm with upd
				// Mark this ItemMenuNum as processed
				processed[ordItm.ItemMenuNum] = true
			}
		}
	}

	for _, updn := range update.MenuIndications {
		if !processed[updn.ItemMenuNum] {
			c.OrderItems.MenuIndications = append(c.OrderItems.MenuIndications, updn)
		}
	}

	c.cleanOrderItems()

	return nil
}

func (c *CustomerOrder) UpdateOrInsertCurrentOrder(db *sql.DB, senderNum string, catalogueID string, update OrderItems, isAutoInc bool) error {
	// Try to find the order in the database
	err := c.SetCurrentOrderFromDB(db, senderNum, isAutoInc)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			err := c.insertOrder(db)
			if err != nil {
				log.Printf("error inserting the order in the DB: %v", err)
				return err
			}
		} else {
			// Some other error occurred
			return err
		}
	} else {

		err = c.UpdateCustOrdItems(update)
		if err != nil {
			log.Printf("error writing the new values to the current order: %v", err)
			return err
		}
		// Keep in mind This will return without errors if the row does not exist
		err = c.updateCurrentOrder(db)
		if err != nil {
			log.Printf("error updating the order in the DB: %v", err)
			return err
		}
	}

	return nil
}

func (c *CustomerOrder) BuildItemName(itemNamePrefix string) string {
	return itemNamePrefix + strconv.Itoa(c.OrderID)
}

// Main function to tally the order
func (c *CustomerOrder) TallyOrder(db *sql.DB, senderNum string, isAutoInc bool) (int, string, error) {
	isInited := c.checkInitialization(db, senderNum, isAutoInc)
	if isInited != custOrderInitState {
		return -1, "", fmt.Errorf("while tallying the order, no current order")
	}

	cartTotal, cartSummary := c.OrderItems.CalculatePrice()
	return cartTotal, cartSummary, nil
}
