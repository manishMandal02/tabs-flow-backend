package e2e_test

import (
	"fmt"
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

func (s *UserFlowTestSuite) TestUserRegisterFlow() {

	reqBody := fmt.Sprintf(`{
		"email": "%s",
		}`, TestUser.Email)

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, s.Config.ApiDomainName+"/auth/send-otp", nil, []byte(reqBody), http.DefaultClient)

	s.NoError(err)
	s.Equal(res.StatusCode, 200)

	// check sqs message for otp event

	// check db (session table) for otp
}
