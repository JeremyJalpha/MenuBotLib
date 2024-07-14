package menubotlib

// CatalogueSelection represents a section of the catalogue with a specific pricing regime
type CatalogueSelection struct {
	Preamble string
	Items    []CatalogueItem
}

// Iterate over Questions array and populate questions array
func (s *CatalogueSelection) CatalogueSelectionAsAString() string {
	allItems := s.Preamble + "\n"

	for _, item := range s.Items {
		allItems += item.CatalogueItemAsAString()
	}

	return allItems
}

// Iterates over CatalogueSelection and returns the concatted string
func AssembleCatalogueSelections(pricelistpreamble string, ctlgselections []CatalogueSelection) string {
	selectionString := pricelistpreamble + "\n\n"

	for i, selection := range ctlgselections {
		selectionString += selection.CatalogueSelectionAsAString()
		if i < len(ctlgselections)-1 {
			selectionString += "\n"
		}
	}

	return selectionString
}

func CmpsCtlgSlctnsFromCtlgItms(ctlgitems []CatalogueItem) []CatalogueSelection {
	var selections []CatalogueSelection
	var currentSelection CatalogueSelection

	for _, item := range ctlgitems {
		if currentSelection.Preamble == "" {
			// Initialize the first selection
			currentSelection.Preamble = item.Selection
		} else if currentSelection.Preamble != item.Selection {
			// Start a new selection
			selections = append(selections, currentSelection)
			currentSelection = CatalogueSelection{
				Preamble: item.Selection,
			}
		}

		// Add the item to the current selection
		currentSelection.Items = append(currentSelection.Items, item)
	}

	// Add the last selection (if any)
	if currentSelection.Preamble != "" {
		selections = append(selections, currentSelection)
	}

	return selections
}
