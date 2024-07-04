package menubotlib

type CheckoutInfo struct {
	ReturnURL      string
	CancelURL      string
	NotifyURL      string
	MerchantId     string
	MerchantKey    string
	Passphrase     string
	HostURL        string
	ItemNamePrefix string
}
