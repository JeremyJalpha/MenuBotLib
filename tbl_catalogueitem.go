package menubotlib

import (
	"database/sql"
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

// GetCatalogueItemsFromDB retrieves catalogue items from the database based on catalogueID.
func GetCatalogueItemsFromDB(db *sql.DB, catalogueID string) ([]CatalogueItem, error) {
	queryString := `SELECT catalogueID, catalogueitemID, "selection", "item", "options", pricingType FROM catalogueitem WHERE catalogueID = $1`
	rows, err := db.Query(queryString, catalogueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CatalogueItem
	for rows.Next() {
		var item CatalogueItem
		err := rows.Scan(&item.CatalogueItemID, &item.Selection, &item.Item, &item.Options, &item.PricingType)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
