package e2e_tests

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// run e2e tests
func TestE2E(t *testing.T) {
	suite.Run(t, new(UserAuthSuite))
	suite.Run(t, new(SpaceSuite))
	suite.Run(t, new(NotesSuite))
	suite.Run(t, new(NotificationSuite))
}
