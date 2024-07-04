package menubotlib

import (
	"fmt"
)

const (
	// TODO: Add catalogueID selection method.
	catalogueID string = "WeAreGettingThePig"

	toolsPreamble = "Tetinus sold seperately"

	gardeningSelectionPreamble = "Gardening:"
	kitchenSelectionPreamble   = "Kitchen:"
	diySelectionPreamble       = "DIY:"
	techSelectionPreamble      = "Tech:"
	booksSelectionPreamble     = "Books:"
)

// CatalogueSelection represents a section of the catalogue with a specific pricing regime
type CatalogueSelection struct {
	Preamble string
	Items    []CatalogueItem
}

var gardeningSelection = CatalogueSelection{
	Preamble: gardeningSelectionPreamble,
	Items: []CatalogueItem{
		{
			CatalogueItemID: 1,
			Item:            "Rusty Garden Spade",
			Options: []string{
				"Gold plated",
				"Chrome plated",
				"Wrought iron",
			},
			PricingType: SingleItem,
		},
		{
			CatalogueItemID: 2,
			Item:            "Bent Garden fork",
			Options: []string{
				"With handle",
				"With pierced shrew",
			},
		},
		{
			CatalogueItemID: 3,
			Item:            "Bristless Garden Broom",
			Options: []string{
				"Roomba version - Vacuum not included",
				"Handleless bristled version",
			},
		},
	},
}

var kitchenSelection = CatalogueSelection{
	Preamble: kitchenSelectionPreamble,
	Items: []CatalogueItem{
		{
			CatalogueItemID: 7,
			Item:            "Microwave",
			Options: []string{
				"5L",
				"7L",
			},
			PricingType: SingleItem,
		},
	},
}

var diySelection = CatalogueSelection{
	Preamble: diySelectionPreamble,
	Items: []CatalogueItem{
		{
			CatalogueItemID: 10,
			Item:            "Drill",
			Options: []string{
				"Ryobi",
			},
			PricingType: SingleItem,
		},
	},
}

var techSelection = CatalogueSelection{
	Preamble: techSelectionPreamble,
	Items: []CatalogueItem{
		{
			CatalogueItemID: 11,
			Item:            "Laptop",
			Options:         []string{},
			PricingType:     SingleItem,
		},
	},
}

var bookSelection = CatalogueSelection{
	Preamble: techSelectionPreamble,
	Items: []CatalogueItem{
		{
			CatalogueItemID: 13,
			Item:            "Lord of the flies",
			Options: []string{
				"paperback",
			},
			PricingType: SingleItem,
		},
	},
}

var selections = []CatalogueSelection{
	gardeningSelection,
	kitchenSelection,
	diySelection,
	techSelection,
	bookSelection,
}

// Generate a string for a single question and answer
func CatalogueItemAsAString(item CatalogueItem) string {
	optionsText := ""
	for i, option := range item.Options {
		optionsText += fmt.Sprintf("   %d. %s\n", i+1, option)
	}

	qA := fmt.Sprintf("%d: %s\n%s\n", item.CatalogueItemID, item.Item, optionsText)

	return qA
}

// Iterate over Questions array and populate questions array
func SingleSelectionAsAString(section CatalogueSelection) string {
	allItems := section.Preamble + "\n"

	for _, item := range section.Items {
		allItems += CatalogueItemAsAString(item)
	}

	return allItems
}

// Updated function to use the new CatalogueSection structure
func PriceListAsAString() string {
	selectionString := toolsPreamble + "\n\n"

	// Outdoor Selection
	selectionString += SingleSelectionAsAString(gardeningSelection) + "\n"

	// Indoor Selection (assuming it's defined elsewhere in your code)
	selectionString += SingleSelectionAsAString(kitchenSelection) + "\n"

	// Vape Selection
	selectionString += SingleSelectionAsAString(diySelection) + "\n"

	// Joints Selection
	selectionString += SingleSelectionAsAString(techSelection) + "\n"

	// Edibles Selection
	selectionString += SingleSelectionAsAString(bookSelection)

	return selectionString
}
