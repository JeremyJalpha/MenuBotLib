package menubotlib

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"database/sql"

	_ "github.com/lib/pq"
)

const (
	custOrderInitState = "Initialized"
	whatsAppServer     = "s.whatsapp.net"

	sayMenu = "For a command list please type & send-: menu?\nPlease include the question mark."

	reminderGreeting = "Please save your email address, by typing & sending-: update email: example@emailprovider.com"

	coldGreeting = "Hello there, I don't believe we've met before."

	smartyPantsGreeting = "Hey there smarty pants, I see you've been here before."

	noCommandText = "Err:NC, Sorry I couldn't identify a command in your mesasge."

	unhandledCommandException = "Err:CF, Something went wrong processing your request."

	newSightingAlternate = "Or use the web form to log a new baboon sighting: "

	newSightingPostURL = "https://baboonobsbot-b6ee90798e7e.herokuapp.com/newsighting"

	newSightingCommand = "new sighting: [address], tribe size, activity level, [current activity]"

	newSightingSqrBrkts = "_(with square brackets)_"

	updateOrderCommand = "update order X:newAmount"

	UpdateOrderCommExpl = "Where X is the item's price list number, item order not important."

	fullOrderExample = `An order of: 
1 Tshirt, 
2 hats, 
1 Mug.

Should look like-: update order 1:1, 2:2, 3:1`

	deleteOrder = "To remove an item from your order, use-: update order X:0"

	shopComands = "to save your order please type & send-:" + updateOrderCommand + "\n" + UpdateOrderCommExpl +
		"\n\n" + fullOrderExample +
		"\n\n" + deleteOrder +
		"\n\n" + "currentorder? - Prints your current pending order." +
		"\n" + "To checkout type & send-: checkoutnow?"

	queryCommands = `menu? - Prints this menu.
shop? - Prints the shop price list.
userinfo? - Prints your user info.`

	updateCommands = `update email: newEmail
update nickname: newNickname
update consent: newConsent`

	mainMenu = "Main Menu, command list:" +
		"\n\n" + newSightingCommand + "\n" + newSightingSqrBrkts +
		"\n\n" + newSightingAlternate + "\n" + newSightingPostURL +
		"\n\n" + queryCommands +
		"\n\n" + updateCommands

	prclstPreamble = "Welcome to B.O.B's Shop," +
		"\n\n" + shopComands
)

type Command interface {
	Execute(db *sql.DB, convo *ConversationContext, isAutoInc bool) error
}

type CommandCollection []Command

type UpdateUserInfoCommand struct {
	Name string
	Text string
}

type UpdateOrderCommand struct {
	Text string
}

type NewSightingCommand struct {
	Text string
}

type QuestionCommand struct {
	Name string
	Text string
}

func (cmd UpdateUserInfoCommand) Execute(db *sql.DB, convo *ConversationContext, isAutoInc bool) error {
	var colName = strings.TrimSpace(strings.TrimPrefix(cmd.Name, "update"))
	err := convo.UserInfo.UpdateSingularUserInfoField(db, colName, cmd.Text)
	if err != nil {
		return fmt.Errorf("unhandled error updating user info: %v", err)
	}
	return errors.New("successfully updated user info." + colName + " to " + cmd.Text)
}

func (cmd UpdateOrderCommand) Execute(db *sql.DB, convo *ConversationContext, isAutoInc bool) error {
	updates, err := ParseUpdateOrderCommand(cmd.Text)
	if err != nil {
		return fmt.Errorf("error parsing update answers command: %v", err)
	}

	err = convo.CurrentOrder.UpdateOrInsertCurrentOrder(db, convo.UserInfo.CellNumber, OrderItems{MenuIndications: updates}, isAutoInc)
	if err != nil {
		return fmt.Errorf("unhandled error updating order: %v", err)
	}
	return errors.New("successfully updated current order")
}

func (cmd NewSightingCommand) Execute(db *sql.DB, convo *ConversationContext, isAutoInc bool) error {
	sighting, err := cmd.ParseNewSightingCommand(convo.UserInfo)
	if err != nil {
		return fmt.Errorf("error parsing new sighting: %v", err)
	}

	err = AddSighting(db, sighting, convo.UserInfo)
	if err != nil {
		return fmt.Errorf("unhandled error adding new sighting: %v", err)
	}
	return errors.New("successfully added new sighting")
}

func (cmd QuestionCommand) Execute(db *sql.DB, convo *ConversationContext, isAutoInc bool) error {
	return fmt.Errorf("%s", cmd.Text)
}

func BeginCheckout(db *sql.DB, ui UserInfo, ctlgselections []CatalogueSelection, c CustomerOrder, checkoutUrls CheckoutInfo, isAutoInc bool) string {

	// Create a new URL object for each URL
	returnURL, _ := url.Parse(checkoutUrls.ReturnURL)
	cancelURL, _ := url.Parse(checkoutUrls.CancelURL)
	notifyURL, _ := url.Parse(checkoutUrls.NotifyURL)

	// Initialize checkoutURLs with the new URLs
	checkoutUrls.ReturnURL = returnURL.String()
	checkoutUrls.CancelURL = cancelURL.String()
	checkoutUrls.NotifyURL = notifyURL.String()

	//Tally the order and then create a CheckoutCart struct
	cartTotal, cartSummary, err := c.TallyOrder(db, ui.CellNumber, ctlgselections, isAutoInc)
	if err != nil {
		return err.Error()
	}
	cart := CheckoutCart{
		ItemName:      c.BuildItemName(checkoutUrls.ItemNamePrefix),
		CartTotal:     cartTotal,
		OrderID:       c.OrderID,
		CustFirstName: ui.NickName.String,
		CustLastName:  ui.CellNumber,
		CustEmail:     ui.Email.String}
	return cartSummary + "/n/n" + ProcessPayment(cart, checkoutUrls)
}

func parseQuestionCommand(match string, db *sql.DB, convo *ConversationContext, checkoutUrls CheckoutInfo, isAutoInc bool) Command {
	switch match {
	case "currentorder?":
		return QuestionCommand{Text: convo.CurrentOrder.GetCurrentOrderAsAString(db, convo.UserInfo.CellNumber, isAutoInc)}
	case "shop?":
		return QuestionCommand{Text: prclstPreamble + "\n\n" + AssembleCatalogueSelections(convo.Pricelist.PrlstPreamble, convo.Pricelist.Catalogue)}
	case "userinfo?":
		return QuestionCommand{Text: convo.UserInfo.GetUserInfoAsAString()}
	case "checkoutnow?":
		return QuestionCommand{Text: BeginCheckout(db, convo.UserInfo, convo.Pricelist.Catalogue, convo.CurrentOrder, checkoutUrls, isAutoInc)}
	default:
		return QuestionCommand{Text: mainMenu}
	}
}

func (cc CommandCollection) ProcessCommands(convo *ConversationContext, db *sql.DB, isAutoInc bool) string {
	var errors []string
	for _, command := range cc {
		err := command.Execute(db, convo, isAutoInc)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}
	return strings.Join(errors, "\n")
}

func GetResponseToMsg(convo *ConversationContext, db *sql.DB, checkoutUrls CheckoutInfo, isAutoInc bool) string {
	commandRes := unhandledCommandException
	commands := GetCommandsFromLastMessage(convo.MessageBody, convo, db, checkoutUrls, isAutoInc)
	if len(commands) != 0 {
		// Process commands
		commandRes_Temp := CommandCollection(commands).ProcessCommands(convo, db, isAutoInc)
		if commandRes_Temp != "" && commandRes_Temp != " " && commandRes_Temp != "\n" {
			commandRes = commandRes_Temp
		}
	} else {
		commandRes = noCommandText
	}

	if !convo.UserExisted {
		if commandRes != noCommandText {
			commandRes = smartyPantsGreeting + "\n\n" + commandRes + "\n\n" + reminderGreeting + "\n\n" + sayMenu
		} else {
			commandRes = coldGreeting + "\n\n" + reminderGreeting + "\n\n" + sayMenu
		}
	} else if commandRes == noCommandText {
		commandRes += "\n\n" + sayMenu
	}

	convo.UserExisted = true

	return commandRes
}

// Precompile regular expressions
var (
	regexQuestionMark  = regexp.MustCompile(`(menu\?|shop\?|userinfo\?|currentorder\?|checkoutnow\?)`)
	regexUpdateField   = regexp.MustCompile(`(update email|update nickname|update social|update consent):\s*(\S*)`)
	regexUpdateAnswers = regexp.MustCompile(`(update order):?\s*(.*)`)
	regexNewSighting   = regexp.MustCompile(`(new sighting):?\s*(.*)`)
)

func GetCommandsFromLastMessage(messageBody string, convo *ConversationContext, db *sql.DB, checkoutUrls CheckoutInfo, isAutoInc bool) []Command {
	var commands []Command
	messageBody = strings.ToLower(messageBody)

	// Use precompiled regular expressions
	if matches := regexQuestionMark.FindAllStringSubmatch(messageBody, -1); matches != nil {
		for _, match := range matches {
			commands = append(commands, parseQuestionCommand(match[1], db, convo, checkoutUrls, isAutoInc))
		}
	}

	if matches := regexUpdateField.FindAllStringSubmatch(messageBody, -1); matches != nil {
		for _, match := range matches {
			commands = append(commands, UpdateUserInfoCommand{Name: match[1], Text: match[2]})
		}
	}

	if matches := regexUpdateAnswers.FindAllStringSubmatch(messageBody, -1); matches != nil {
		for _, match := range matches {
			commands = append(commands, UpdateOrderCommand{Text: match[2]})
		}
	}

	if matches := regexNewSighting.FindAllStringSubmatch(messageBody, -1); matches != nil {
		for _, match := range matches {
			commands = append(commands, NewSightingCommand{Text: match[2]})
		}
	}

	return commands
}

// finds the index of the first comma after the last closing bracket
func findFirstClsngBrkt(input string) int {
	foundBracket := false
	for i, char := range input {
		if char == ']' {
			foundBracket = true
		}
		if foundBracket && char == ',' {
			return i
		}
	}
	return -1 // Return -1 if no such comma is found
}

// finds the index of the first opening bracket
func findFirstOpnngBrkt(input string) int {
	for i, char := range input {
		if char == '[' {
			return i
		}
	}
	return -1 // Return -1 if no such bracket is found
}

// finds the index of the first comma after the last opening bracket
func findLastOpnngBrkt(input string) int {
	foundBracket := false
	for i, char := range input {
		if char == ',' {
			foundBracket = true
		}
		if foundBracket && char == '[' {
			return i
		}
	}
	return -1 // Return -1 if no such comma is found
}

// finds the index of the last closing bracket
func findLastClsngBrkt(input string) int {
	for i, char := range input {
		if char == ']' {
			return i
		}
	}
	return -1 // Return -1 if no such bracket is found
}

// parses the command text to extract the content between brackets and modifies the original string
func parseAddressBrktGroup(commandText *string) (string, error) {
	clB := findFirstClsngBrkt(*commandText)
	if clB == -1 {
		return "", errors.New("pattern '],' not found")
	}
	opB := findFirstOpnngBrkt(*commandText)
	if clB < opB {
		return "", errors.New("] found before [")
	}

	// Extract the substring between the brackets
	bracketContent := (*commandText)[opB+1 : clB-1]

	// Trim the brackets from the original string
	*commandText = (*commandText)[:opB] + (*commandText)[clB+1:]

	return bracketContent, nil
}

// parses the command text to extract the content between brackets and modifies the original string
func parseCurrentActionBrktGroup(commandText *string) (string, error) {
	opB := findLastOpnngBrkt(*commandText)
	if opB == -1 {
		return "", errors.New("pattern ',[' not found")
	}
	clB := findLastClsngBrkt(*commandText)
	if opB > clB {
		return "", errors.New("] found before [")
	}

	// Extract the substring between the brackets
	bracketContent := (*commandText)[opB+1 : clB]

	// Trim the brackets from the original string
	*commandText = (*commandText)[:opB] + (*commandText)[clB+1:]

	return bracketContent, nil
}

func validateActivityLevel(level string) (ActivityLevel, error) {
	switch level {
	case "1", "low":
		return T_Low, nil
	case "2", "medium":
		return T_Medium, nil
	case "3", "high":
		return T_High, nil
	default:
		return "", errors.New("invalid activity level")
	}
}

func validateTribeSize(size string) (TribeSize, error) {
	switch size {
	case "1", "small":
		return S_Small, nil
	case "2", "medium":
		return S_Medium, nil
	case "3", "large":
		return S_Large, nil
	default:
		return "", errors.New("invalid tribe size")
	}
}

func (cmd NewSightingCommand) ParseNewSightingCommand(u UserInfo) (Sighting, error) {

	sghtngAddress, err := parseAddressBrktGroup(&cmd.Text)
	if err != nil {
		return Sighting{}, fmt.Errorf("failed to parse sighting text: %v", err)
	}

	curActivity, err := parseCurrentActionBrktGroup(&cmd.Text)
	if err != nil {
		return Sighting{}, fmt.Errorf("failed to parse sighting text: %v", err)
	}

	cmd.Text = strings.TrimSpace(cmd.Text)
	cmd.Text = strings.TrimPrefix(cmd.Text, ",")
	cmd.Text = strings.TrimSuffix(cmd.Text, ",")

	remainingFields := strings.Split(cmd.Text, ",")
	if len(remainingFields) != 2 {
		return Sighting{}, errors.New("too many or too few commas provided in sighting text")
	}

	TribeSize, err := validateTribeSize(strings.TrimSpace(remainingFields[0]))
	if err != nil {
		return Sighting{}, fmt.Errorf("failed to convert inputted tribe size: %v", err)
	}
	ActivityLevel, err := validateActivityLevel(strings.TrimSpace(remainingFields[1]))
	if err != nil {
		return Sighting{}, fmt.Errorf("failed to convert inputted activity level: %v", err)
	}

	return Sighting{
		CellNumber:      u.CellNumber,
		SightingAddress: sghtngAddress,
		CurrentActivity: curActivity,
		TribeSize:       TribeSize,
		ActivityLevel:   ActivityLevel,
		DateTimePosted:  sql.NullTime{Time: time.Now(), Valid: true},
	}, nil
}

func ParseUpdateOrderCommand(commandText string) ([]MenuIndication, error) {
	// Remove "update order" prefix
	commandText = strings.TrimPrefix(commandText, "update order")
	commandText = strings.TrimPrefix(commandText, ":")
	commandText = strings.TrimSpace(commandText)
	commandText = strings.Replace(commandText, " ", "", 1)

	// Regular expression to match "ItemMenuNum: ItemAmount" pairs
	re := regexp.MustCompile(`\b\d+:\s*(?:\d+x\d+(?:,\s*)?)+`)

	// Find all matches in the commandText
	matches := re.FindAllString(commandText, -1)

	// Remove matched parts from the commandText
	for k, match := range matches {
		trimmedMatch := strings.TrimSpace(match)
		trimmedMatch = strings.TrimSuffix(trimmedMatch, ",")
		matches[k] = trimmedMatch
		commandText = strings.Replace(commandText, match, "", 1)
	}

	// Trim any remaining whitespace or commas
	commandText = strings.Trim(commandText, ",")

	// Initialize slice to store OrderItems
	var orderItems []MenuIndication

	// Process each match
	for _, match := range matches {
		orderItem, err := parseOrderItem(match)
		if err != nil {
			return nil, err
		}
		orderItems = append(orderItems, orderItem)
	}

	// Process remaining commandText for simple "ItemMenuNum: ItemAmount" pairs
	if commandText != "" {
		remainingItems := strings.Split(commandText, ",")
		for _, item := range remainingItems {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			orderItem, err := parseOrderItem(item)
			if err != nil {
				return nil, err
			}
			orderItems = append(orderItems, orderItem)
		}
	}

	return orderItems, nil
}

func parseOrderItem(item string) (MenuIndication, error) {
	parts := strings.SplitN(item, ":", 2)
	if len(parts) != 2 {
		return MenuIndication{}, fmt.Errorf("failed to parse item: %s", item)
	}

	// Parse ItemMenuNum
	itemMenuNum, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return MenuIndication{}, fmt.Errorf("failed to parse ItemMenuNum: %v", err)
	}

	// Trim and clean up ItemAmount
	itemAmount := strings.TrimSpace(parts[1])
	itemAmount = strings.TrimSuffix(itemAmount, ",")

	return MenuIndication{
		ItemMenuNum: itemMenuNum,
		ItemAmount:  itemAmount,
	}, nil
}
