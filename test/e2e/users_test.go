package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
	"github.com/stretchr/testify/suite"
)

func (s *UserFlowTestSuite) TestUserRegisterFlow() {

	updateLambdaEventSourceMappingState(s.LambdaClient, s.SQSLambdaMapping.EmailService, true)

	time.Sleep(1 * time.Second)

	// send otp to email
	reqBody := fmt.Sprintf(`{
		"email": "%s"
		}`, TestUser.Email)

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/auth/send-otp", nil, []byte(reqBody), http.DefaultClient)

	s.NoError(err)

	if err != nil {
		s.FailNow(err.Error())
	}

	s.Equal(http.StatusOK, res.StatusCode)

	logger.Info("OTP sent to email")

	// check sqs message for otp event
	time.Sleep(1 * time.Second)

	otpMessage, err := getSQSQueueMessage[events.SendOTPPayload](s.SQSClient, s.ENV.EmailQueueURL)

	s.NoError(err)

	if err != nil {
		s.FailNow(err.Error())
	}

	s.Equal(otpMessage.Payload.Email, TestUser.Email)

	otp := otpMessage.Payload.OTP

	updateLambdaEventSourceMappingState(s.LambdaClient, s.SQSLambdaMapping.EmailService, false)

	logger.Info("OTP: %v", otp)

	// check db (session table) for otp
	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: TestUser.Email},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY_SESSIONS.OTP(otp)},
	}

	item, err := s.DDBClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(s.ENV.SessionTable),
		Key:       key,
	})

	s.NoError(err)

	s.NotNil(item)

	// verify otp
	reqBody = fmt.Sprintf(`{
		"email": "%s",
		"otp": "%s"
		}`, TestUser.Email, otp)

	res, respBody, err := utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/auth/verify-otp", nil, []byte(reqBody), http.DefaultClient)

	s.NoError(err)

	s.Equal(res.StatusCode, 200)

	// res body

	resData := struct {
		UserId  string `json:"userId"`
		NewUser bool   `json:"isNewUser"`
	}{}

	err = json.Unmarshal([]byte(respBody), &resData)

	s.NoError(err)

	s.True(resData.NewUser, "user should be new")

	for _, c := range res.Cookies() {
		if c.Name == "access_token" {
			s.SessionCookie = c
		}
	}

	logger.Dev("cookiejar: %v", s.SessionCookie)

	// create user

	reqBody = fmt.Sprintf(`{
		"id": "%s",
		"fullName": "%s",
		"email": "%s",
		"profilePic": "%s"
		}`,
		resData.UserId, TestUser.FullName, TestUser.Email, TestUser.ProfilePic)

	reqHeader := map[string]string{
		"Cookie": s.SessionCookie.String(),
	}

	res, _, err = utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/users", reqHeader, []byte(reqBody), http.DefaultClient)

	s.NoError(err)
	s.Equal(200, res.StatusCode)

	logger.Info("User created")

}

// run e2e tests
func TestUserFlowTestSuite(t *testing.T) {
	suite.Run(t, new(UserFlowTestSuite))
}
