package menubotlib

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// Define a custom type for PricingType
type PricingType string

// Define constants for the PricingType values
const (
	WeightItem PricingType = "WeightItem"
	SingleItem PricingType = "SingleItem"
)

type CatalogueItem struct {
	CatalogueID     string
	CatalogueItemID int
	Selection       string
	Item            string
	Options         []string
	PricingType     PricingType
}

// Generate a string for a single question and answer
func (i *CatalogueItem) CatalogueItemAsAString() string {
	optionsText := ""
	for i, option := range i.Options {
		optionsText += fmt.Sprintf("   %d. %s\n", i+1, option)
	}

	qA := fmt.Sprintf("%d: %s\n%s\n", i.CatalogueItemID, i.Item, optionsText)

	return qA
}

func InsertCatalogueItems(db *sql.DB, selections []CatalogueSelection) error {
	insertStmt := `
	INSERT INTO catalogueitem (catalogueID, catalogueitemID, "selection", "item", "options", pricingType)
	VALUES (?, ?, ?, ?, ?, ?);`

	for _, selection := range selections {
		for _, item := range selection.Items {
			// Marshal the []string into JSON
			optionsJSON, err := json.Marshal(item.Options)
			if err != nil {
				return err
			}
			_, err = db.Exec(insertStmt, item.CatalogueID, item.CatalogueItemID, selection.Preamble, item.Item, optionsJSON, item.PricingType)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func GetCatalogueItemsFromDB(db *sql.DB, catalogueid string) ([]CatalogueItem, error) {
	query := `
	SELECT catalogueID, catalogueitemID, "selection", "item", "options", pricingType
	FROM catalogueitem
	WHERE catalogueID = $1;`

	rows, err := db.Query(query, catalogueid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rtnItems []CatalogueItem

	// Example:
	// {"MenuIndications":[
	// {"ItemMenuNum":1,"ItemAmount":"12"},
	// {"ItemMenuNum":14,"ItemAmount":"10"}]}

	for rows.Next() {
		var item CatalogueItem
		var optionsStr string

		err := rows.Scan(&item.CatalogueID, &item.CatalogueItemID, &item.Selection, &item.Item, &optionsStr, &item.PricingType)
		if err != nil {
			item = CatalogueItem{}
		}

		// Unmarshal the JSON back into a []string
		var options []string
		err = json.Unmarshal([]byte(optionsStr), &options)
		if err != nil {
			item.Options = nil
		} else {
			item.Options = options
		}

		rtnItems = append(rtnItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rtnItems, nil
}
