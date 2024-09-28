package email

import (
	"encoding/json"
	"errors"
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

type NameAddr struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type ZeptoMailBody struct {
	TemplateKey string     `json:"template_key"`
	To          []NameAddr `json:"to"`
	From        *NameAddr  `json:"from"`
}

type otpMergeInfo struct {
	OTP string `json:"OTP"`
}
type otpEmailBody struct {
	*ZeptoMailBody
	MergeInfo *otpMergeInfo `json:"merge_info"`
}

type ZeptoMail struct {
	URL     string
	APIKey  string
	Headers map[string]string
	From    *NameAddr
}

func NewZeptoMail() *ZeptoMail {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Zoho-enczapikey " + config.ZEPTO_MAIL_API_KEY,
	}

	return &ZeptoMail{
		URL:     config.ZEPTO_MAIL_API_URL,
		APIKey:  config.ZEPTO_MAIL_API_KEY,
		Headers: headers,
		From: &NameAddr{
			Name:    "TabsFlow Support",
			Address: "support@tabsflow.com",
		},
	}
}

func (z *ZeptoMail) SendOTPMail(otp string, to *NameAddr) error {

	body := &otpEmailBody{
		ZeptoMailBody: &ZeptoMailBody{
			TemplateKey: ZeptoMailTemplates["otp"],
			To:          append([]NameAddr{}, *to),
			From: &NameAddr{
				Name:    z.From.Name,
				Address: z.From.Address,
			},
		},
		MergeInfo: &otpMergeInfo{
			OTP: otp,
		},
	}

	bodyBytes, err := json.Marshal(body)

	if err != nil {
		return err
	}

	err = sendMail(z.URL, z.Headers, bodyBytes)

	if err != nil {
		return err
	}

	return nil
}

type welcomeMergeInfo struct {
	Name         string `json:"name"`
	TrailEndDate string `json:"trail_end_date"`
	TrailEndLink string `json:"trail_end_link"`
}
type welcomeEmailBody struct {
	*ZeptoMailBody
	MergeInfo *welcomeMergeInfo `json:"merge_info"`
}

func (z *ZeptoMail) sendWelcomeMail(to *NameAddr, trailEndDate string) error {

	body := &welcomeEmailBody{
		ZeptoMailBody: &ZeptoMailBody{
			TemplateKey: ZeptoMailTemplates["opt"],
			To:          append([]NameAddr{}, *to),
			From: &NameAddr{
				Name:    z.From.Name,
				Address: z.From.Address,
			},
		},
		MergeInfo: &welcomeMergeInfo{
			Name:         to.Name,
			TrailEndDate: trailEndDate,
			TrailEndLink: "https://tabsflow.com/",
		},
	}

	bodyBytes, err := json.Marshal(body)

	if err != nil {
		return err
	}
	err = sendMail(z.URL, z.Headers, bodyBytes)

	if err != nil {
		return err
	}

	return nil
}

// helper
func sendMail(url string, headers map[string]string, body []byte) error {
	res, err := utils.MakeHTTPRequest(http.MethodPost, url, headers, body)

	if err != nil {
		logger.Error(fmt.Sprintf("[email_service] Error sending email email_body:", string(body)), err)
		return err
	}

	defer res.Body.Close()

	var resBody interface{}

	json.NewDecoder(res.Body).Decode(&resBody)

	fmt.Println("response body:: ", resBody.(map[string]interface{}))

	logger.Dev(fmt.Sprintf("ZeptoMail Response code: %v,", res.StatusCode))

	if !strings.HasPrefix(string(res.StatusCode), "2") {
		logger.Error("[email_service] unsuccessful response from zepto mail.", errors.New(res.Status))
		return errors.New("unsuccessful response from zepto mail")
	}
	return nil
}
