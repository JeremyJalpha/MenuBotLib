package menubotlib

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

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
		"\n\n" + queryCommands +
		"\n\n" + updateCommands

	prclstPreamble = "Welcome to the Shop," +
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
	regexQuestionMark  = regexp.MustCompile(`(menu\?|fr\.prlist\?|userinfo\?|currentorder\?|checkoutnow\?)`)
	regexUpdateField   = regexp.MustCompile(`(update email|update nickname|update social|update consent):\s*(\S*)`)
	regexUpdateAnswers = regexp.MustCompile(`(update order):?\s*(.*)`)
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

	return commands
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
