package main

import (
	"encoding/json"
	"time"
)

// moscowLoc — часовой пояс Europe/Moscow (UTC+3, Россия не переходит на зимнее время с 2014 года).
var moscowLoc = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.FixedZone("MSK", 3*60*60)
	}
	return loc
}()

const dateTimeLayout = "2006-01-02T15:04:05"

// DateTime — время с кастомным парсингом JSON в часовом поясе Europe/Moscow.
type DateTime struct {
	time.Time
}

// UnmarshalJSON десериализует строку формата "2006-01-02T15:04:05" в часовом поясе Europe/Moscow.
func (d *DateTime) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	t, err := time.ParseInLocation(dateTimeLayout, s, moscowLoc)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

// MarshalJSON сериализует время в строку формата "2006-01-02T15:04:05" по Europe/Moscow.
func (d DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.In(moscowLoc).Format(dateTimeLayout))
}

type Receipt struct {
	Key           string    `json:"key"`
	CreatedDate   DateTime  `json:"createdDate"`
	ReceiveDate   DateTime  `json:"receiveDate"`
	TotalSum      string    `json:"totalSum"`
	FiscalNumber  string    `json:"fiscalDriveNumber"`
	DocumentNum   string    `json:"fiscalDocumentNumber"`
	FiscalSign    string    `json:"fiscalSign"`
	OperationType int       `json:"operationType"`
}

type ReceiptResponse struct {
	Receipts []Receipt `json:"receipts"`
	HasMore  bool      `json:"hasMore"`
}

type Item struct {
	Name        string  `json:"name"`
	NDS         int     `json:"nds"`
	NDSSum      float64 `json:"ndsSum"`
	PaymentType int     `json:"paymentType"`
	Price       float64 `json:"price"`
	ProductType int     `json:"productType"`
	ProviderINN string  `json:"providerInn"`
	Quantity    float64 `json:"quantity"`
	Sum         float64 `json:"sum"`
}

type FiscalDataResponse struct {
	BuyerPhoneOrAddress     string    `json:"buyerAddress"`
	CashTotalSum            float64   `json:"cashTotalSum"`
	Code                    int       `json:"code"`
	CreditSum               float64   `json:"creditSum"`
	DateTime                DateTime  `json:"dateTime"`
	ECashTotalSum           float64   `json:"ecashTotalSum"`
	FiscalDocumentFormatVer string    `json:"fiscalDocumentFormatVer"`
	FiscalDocumentNumber    int       `json:"fiscalDocumentNumber"`
	FiscalDriveNumber       string    `json:"fiscalDriveNumber"`
	FiscalSign              string    `json:"fiscalSign"`
	Items                   []Item    `json:"items"`
	KKTRegID                string    `json:"kktRegId"`
	NDS10                   float64   `json:"nds10"`
	NDS18                   float64   `json:"nds18"`
	OperationType           int       `json:"operationType"`
	Operator                string    `json:"operator"`
	OperatorINN             string    `json:"operatorInn"`
	PrepaidSum              float64   `json:"prepaidSum"`
	ProvisionSum            float64   `json:"provisionSum"`
	RequestNumber           int       `json:"requestNumber"`
	RetailPlace             string    `json:"retailPlace"`
	RetailPlaceAddress      string    `json:"retailPlaceAddress"`
	ShiftNumber             int       `json:"shiftNumber"`
	TaxationType            int       `json:"taxationType"`
	AppliedTaxationType     int       `json:"appliedTaxationType"`
	TotalSum                float64   `json:"totalSum"`
	User                    string    `json:"user"`
	UserINN                 string    `json:"userInn"`
}

type TransformedReceipt struct {
	ID        string    `json:"_id" bson:"_id"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	Ticket    struct {
		Document struct {
			Receipt FiscalDataResponse `json:"receipt" bson:"receipt"`
		} `json:"document" bson:"document"`
	} `json:"ticket" bson:"ticket"`
}
