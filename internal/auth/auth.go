package auth

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type cognitoLambdaEvent map[string]map[string]interface{}

// TODO - load config from env, and pass to all functions below
// cfg, err := config.LoadDefaultConfig(context.TODO())
// if err != nil {
// 	log.Fatalf("failed to load configuration, %v", err)
// }
// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/

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

	// send email
	queueName := os.Getenv("EMAIL_QUEUE_NAME")

	sqsClient := sqs.New(sqs.Options{})

	urlResult, err := sqsClient.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})

	if err != nil {
		return nil, err
	}

	// TODO - sqs helper fn to send messages

	sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		DelaySeconds: *aws.Int32(1),
		QueueUrl:     urlResult.QueueUrl,
		MessageBody:  aws.String("Send OTP to user"),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Email": {
				DataType:    aws.String("String"),
				StringValue: aws.String("user@example.com"),
			},
			"Event": {
				DataType:    aws.String("String"),
				StringValue: aws.String("OTP"),
			},
			"Payload": {
				DataType:    aws.String("String"),
				StringValue: aws.String(challenge),
			},
		},
	})

	return event, nil
}

// VerifyAuthChallengeResponse validates the user's response to the challenge
func VerifyAuthChallengeResponse(_ context.Context, event cognitoLambdaEvent) (cognitoLambdaEvent, error) {
	expectedAnswer := event["request"]["privateChallengeParameters"].(map[string]string)["answer"]
	userAnswer := event["request"]["challengeAnswer"].(string)
	event["response"]["answerCorrect"] = (expectedAnswer == userAnswer)
	return event, nil
}
