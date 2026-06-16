package mpesa

import (
	"encoding/json"
	"fmt"
	"os"
)

// B2CRequest sends money from the business to a customer phone number.
type B2CRequest struct {
	InitiatorName      string `json:"InitiatorName"`
	SecurityCredential string `json:"SecurityCredential"`
	CommandID          string `json:"CommandID"`
	Amount             int64  `json:"Amount"`
	PartyA             string `json:"PartyA"`
	PartyB             string `json:"PartyB"`
	Remarks            string `json:"Remarks"`
	QueueTimeOutURL    string `json:"QueueTimeOutURL"`
	ResultURL          string `json:"ResultURL"`
	Occasion           string `json:"Occasion"`
}

type B2CResponse struct {
	ConversationID         string `json:"ConversationID"`
	OriginatorConversationID string `json:"OriginatorConversationID"`
	ResponseCode           string `json:"ResponseCode"`
	ResponseDescription    string `json:"ResponseDescription"`
}

// SendB2C initiates a B2C payout to the given phone in KES.
func SendB2C(phone string, amountKES int64, ref string) (*B2CResponse, error) {
	req := B2CRequest{
		InitiatorName:      os.Getenv("MPESA_B2C_INITIATOR_NAME"),
		SecurityCredential: os.Getenv("MPESA_B2C_SECURITY_CREDENTIAL"),
		CommandID:          "BusinessPayment",
		Amount:             amountKES,
		PartyA:             os.Getenv("MPESA_SHORTCODE"),
		PartyB:             phone,
		Remarks:            "YeboBank withdrawal",
		QueueTimeOutURL:    os.Getenv("MPESA_CALLBACK_URL") + "/webhook/mpesa/b2c/timeout",
		ResultURL:          os.Getenv("MPESA_CALLBACK_URL") + "/webhook/mpesa/b2c",
		Occasion:           ref,
	}
	data, err := Default.post("/mpesa/b2c/v3/paymentrequest", req)
	if err != nil {
		return nil, err
	}
	var resp B2CResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("b2c: decode: %w", err)
	}
	if resp.ResponseCode != "0" {
		return nil, fmt.Errorf("b2c: %s", resp.ResponseDescription)
	}
	return &resp, nil
}
