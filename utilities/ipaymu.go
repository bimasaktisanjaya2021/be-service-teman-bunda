package utilities

type IpaymuDirectPaymentResponse struct {
	Status  int
	Message string
	Data    Data
}

type Data struct {
	SessionId     string
	TransactionId int
	ReferenceId   string
	Via           string
	Channel       string
	PaymentNo     string
	PaymentName   string
	Total         float64
	Fee           float64
	Expired       string
}
