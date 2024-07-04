package menubotlib

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type CheckoutCart struct {
	ItemName      string
	CartTotal     int
	CustFirstName string
	CustLastName  string
	CustEmail     string
	OrderID       int
}

type KeyValue struct {
	Key   string
	Value string
}

func escapeParams(params []KeyValue) []KeyValue {
	var escapedParams []KeyValue
	for _, kv := range params {
		escapedParams = append(escapedParams, KeyValue{
			Key:   kv.Key,
			Value: url.QueryEscape(kv.Value),
		})
	}
	return escapedParams
}

func concatParams(params []KeyValue, passPhrase string) string {
	var urlData string
	paramCount := len(params)
	escapedParams := escapeParams(params)

	for i, kv := range escapedParams {
		urlData += kv.Key + "=" + kv.Value
		if i < paramCount-1 {
			urlData += "&"
		}
	}

	// Append passphrase if it exists
	if passPhrase != "" {
		urlData += "&passphrase=" + url.QueryEscape(passPhrase)
	}

	return urlData
}

func generateSignature(data string) string {
	// Generate MD5 hash
	hash := md5.New()
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}

func sliceToValues(params []KeyValue) url.Values {
	values := url.Values{}
	for _, kv := range params {
		values.Set(kv.Key, kv.Value)
	}
	return values
}

func ProcessPayment(cart CheckoutCart, checkoutInfo CheckoutInfo) string {
	params := []KeyValue{
		{"merchant_id", checkoutInfo.MerchantId},
		{"merchant_key", checkoutInfo.MerchantKey},
		{"return_url", checkoutInfo.ReturnURL},
		{"cancel_url", checkoutInfo.CancelURL},
		{"notify_url", checkoutInfo.NotifyURL},
		{"name_first", cart.CustFirstName},
		{"name_last", cart.CustLastName},
		{"email_address", cart.CustEmail},
		{"cell_number", cart.CustLastName},
		{"m_payment_id", strconv.Itoa(cart.OrderID)},
		{"amount", fmt.Sprintf("%.2f", float64(cart.CartTotal))},
		{"item_name", cart.ItemName},
	}

	// Generate the signature
	signature := generateSignature(concatParams(params, checkoutInfo.Passphrase))

	// Convert the map to url.Values
	urlParams := sliceToValues(params)
	urlParams.Add("signature", signature)

	// Make the HTTP POST request
	resp, err := http.PostForm(checkoutInfo.HostURL, urlParams)
	if err != nil {
		fmt.Println("error making POST request:", err)
		return "Checkout initiation failed"
	}
	defer resp.Body.Close()

	// Check if it's a redirect (3xx status code)
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		redirectURL := resp.Header.Get("Location")
		return redirectURL
	}

	return "Checkout initiation failed"
}
