package email

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

var ZeptoMailTemplates = map[string]string{
	"otp":     "2518b.5aadbc61a6c007b3.k1.9b26d2f0-7460-11ef-b8eb-525400ab18e6.191fc46a59f",
	"welcome": "2518b.5aadbc61a6c007b3.k1.98545e20-74c5-11ef-b8eb-525400ab18e6.191fedc7d02",
}

type nameAddr struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type otpMergeInfo struct {
	OTP string `json:"OTP"`
}

type ZeptoMailBody struct {
	TemplateKey string        `json:"template_key"`
	To          []nameAddr    `json:"to"`
	From        *nameAddr     `json:"from"`
	MergeInfo   *otpMergeInfo `json:"merge_info"`
}

func sendOTPMail(otp string, to *nameAddr) error {

	zeptoMailURL := "https://api.zeptomail.com/v1.1/email"

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "zoho-enczapikey " + config.ZEPTO_MAIL_API_KEY,
	}

	body := &ZeptoMailBody{
		TemplateKey: ZeptoMailTemplates["opt"],
		To:          append([]nameAddr{}, *to),
		From: &nameAddr{
			Name:    "TabsFlow",
			Address: "support@tabsflow.com",
		},
		MergeInfo: &otpMergeInfo{
			OTP: otp,
		},
	}

	bodyBytes, err := json.Marshal(body)

	if err != nil {
		return err
	}

	res, err := utils.MakeHTTPRequest(http.MethodPost, zeptoMailURL, headers, bodyBytes)

	if err != nil {
		logger.Error(fmt.Sprintf("[email_service] Error sending OTP to email:", body.From.Address), err)
		return err
	}

	defer res.Body.Close()

	if !strings.HasPrefix(string(res.StatusCode), "2") {
		logger.Error(fmt.Sprintf("[email_service] Error sending OTP to email:", body.From.Address), err)
		return err
	}

	return nil
}
