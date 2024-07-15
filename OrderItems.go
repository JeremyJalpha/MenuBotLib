package menubotlib

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type MenuIndication struct {
	ItemMenuNum int    `json:"ItemMenuNum"`
	ItemAmount  string `json:"ItemAmount"`
}

type OrderItems struct {
	MenuIndications []MenuIndication `json:"MenuIndications"`
}

// Example:
//{"MenuIndications":[{"ItemMenuNum":1,"ItemAmount":"2x3"},{"ItemMenuNum":2,"ItemAmount":"1x5"}]}
//{"MenuIndications":[{"ItemMenuNum":9,"ItemAmount":"12"},{"ItemMenuNum":10,"ItemAmount":"1x3, 3x2, 2x1"},{"ItemMenuNum":6,"ItemAmount":"5"}]}

func findItemInSelections(ItmMnuNum int, ctlgselections []CatalogueSelection) (CatalogueItem, error) {
	for _, selection := range ctlgselections {
		for _, item := range selection.Items {
			if item.CatalogueItemID == ItmMnuNum {
				return item, nil
			}
		}
	}
	return CatalogueItem{}, fmt.Errorf("item menu num not found")
}

func tallyOptions(options []string, userInput string) (int, error) {
	userItems := strings.Split(userInput, ",")
	totalPrice := 0

	for _, userItem := range userItems {
		userItem = strings.TrimSpace(userItem)
		var optionNumber, amount int
		_, err := fmt.Sscanf(userItem, "%dx%d", &optionNumber, &amount)
		if err != nil {
			return -99, fmt.Errorf("while tallying order, error parsing userInput: %s, %v", userItem, err)
		}

		if optionNumber <= 0 || optionNumber > len(options) {
			return -99, fmt.Errorf("while tallying order, invalid option number: %d", optionNumber)
		}

		var price int

		// Define the regular expression pattern
		re := regexp.MustCompile(`@ R(\d+)`)

		// Find the match
		match := re.FindStringSubmatch(options[optionNumber-1])
		if len(match) < 2 {
			return -99, fmt.Errorf("while tallying order, price for item nunmber: " + strconv.Itoa(optionNumber) + " not found in item option string")
		}

		// Convert the extracted string to an integer
		price, err = strconv.Atoi(match[1])
		if err != nil {
			return -99, fmt.Errorf("while tallying order, error parsing option price: %s, %v", options[optionNumber-1], err)
		}

		totalPrice += amount * price
	}

	return totalPrice, nil
}

// Helper function to find the best price based on the order amount and options available
func findBestPrice(orderAmount int, options []string) (int, error) {
	bestPrice := math.MaxInt32
	for _, option := range options {
		var optionWeight int
		var optionPrice int
		_, err := fmt.Sscanf(option, "%dg @ R%d", &optionWeight, &optionPrice)
		if err == nil && orderAmount >= optionWeight {
			if bestPrice == -1 || optionPrice < bestPrice {
				bestPrice = optionPrice
			}
		}
	}
	if bestPrice == math.MaxInt32 {
		// No valid price found
		return -99, fmt.Errorf("while tallying order, error finding best price")
	}
	return bestPrice, nil
}

func (c *OrderItems) CalculatePrice(ctlgselections []CatalogueSelection) (int, string) {
	cartSummary := ""
	cartTotal := 0
	for _, orderItem := range c.MenuIndications {
		// Look up the item in the sections
		foundItem, err := findItemInSelections(orderItem.ItemMenuNum, ctlgselections)
		if err != nil {
			// Add excluded items to the cart summary.
			cartSummary += fmt.Sprintf("while tallying the order, user specified Item menu nunmber: %d not found in price list", orderItem.ItemMenuNum)
			continue
		}
		switch foundItem.PricingType {
		case WeightItem:
			weight, err := strconv.Atoi(orderItem.ItemAmount)
			if err != nil {
				cartSummary += fmt.Sprintf("while tallying the order, error converting userInput to weight: %s to integer: %v", orderItem.ItemAmount, err)
			}
			price, err := findBestPrice(weight, foundItem.Options)
			if err != nil {
				cartSummary += fmt.Sprintln("while tallying the order, error finding best price")
			}
			cartTotal += (weight * price)
		case SingleItem:
			optionsTotal, err := tallyOptions(foundItem.Options, orderItem.ItemAmount)
			if err != nil {
				cartSummary += fmt.Sprintf("while tallying the order, error extracting the order item price: %v", err)
			}
			cartTotal += optionsTotal
		default:
			cartSummary += fmt.Sprintf("while tallying the order, unknown pricing type: %s", foundItem.PricingType)
		}
	}
	return cartTotal, cartSummary
}
