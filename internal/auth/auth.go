package auth

import (
	"context"
)

type cognitoLambdaEvent map[string]map[string]interface{}

// DefineAuthChallenge handles which challenge to present
func DefineAuthChallenge(_ context.Context, event cognitoLambdaEvent) (cognitoLambdaEvent, error) {
	event["response"]["issueTokens"] = false
	event["response"]["failAuthentication"] = false
	event["response"]["challengeName"] = "CUSTOM_CHALLENGE"
	return event, nil
}

// CreateAuthChallenge generates the actual challenge (e.g., OTP)
func CreateAuthChallenge(_ context.Context, event cognitoLambdaEvent) (cognitoLambdaEvent, error) {

	// TODO - check if the user already has a session
	// https://aws.amazon.com/blogs/mobile/implementing-passwordless-email-authentication-with-amazon-cognito/

	challenge := "123456" // Here you'd generate your challenge, e.g., OTP
	event["response"]["publicChallengeParameters"] = map[string]string{"challenge": challenge}
	event["response"]["privateChallengeParameters"] = map[string]string{"answer": challenge}
	event["response"]["challengeMetadata"] = "CUSTOM_CHALLENGE"
	return event, nil
}

// VerifyAuthChallengeResponse validates the user's response to the challenge
func VerifyAuthChallengeResponse(_ context.Context, event cognitoLambdaEvent) (cognitoLambdaEvent, error) {
	expectedAnswer := event["request"]["privateChallengeParameters"].(map[string]string)["answer"]
	userAnswer := event["request"]["challengeAnswer"].(string)
	event["response"]["answerCorrect"] = (expectedAnswer == userAnswer)
	return event, nil
}
