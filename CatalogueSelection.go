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
