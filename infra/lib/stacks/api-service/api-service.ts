import { Construct } from 'constructs';
import * as eventSources from 'aws-cdk-lib/aws-lambda-event-sources';
import { Duration, Stack, StackProps, aws_lambda as lambda } from 'aws-cdk-lib';

import * as sqs from 'aws-cdk-lib/aws-sqs';
import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';

type APIServiceStackProps = StackProps & {
  stage: string;
  appName: string;
  googleClientId: string;
  googleClientSecret: string;
  sesEmail: string;
  emailQueueName: string;
};

export class APIServiceStack extends Stack {
  constructor(scope: Construct, id: string, props: APIServiceStackProps) {
    super(scope, id, props);

    //TODO- email service
    //- sqs queue
    const dlqEmail = new sqs.Queue(this, 'EmailDLQ', {
      visibilityTimeout: Duration.seconds(300)
    });
    const emailQueue = new sqs.Queue(this, props.emailQueueName, {
      queueName: props.emailQueueName,
      visibilityTimeout: Duration.seconds(300),
      deliveryDelay: Duration.seconds(1),
      deadLetterQueue: {
        queue: dlqEmail,
        maxReceiveCount: 3
      }
    });

    //- lambda
    const emailLambda = new GoFunction(this, 'DefineChallenge', {
      entry: 'cmd/email/main.go',
      runtime: lambda.Runtime.PROVIDED_AL2,
      timeout: Duration.seconds(30),
      memorySize: 128
    });

    // grants permissions to lambda
    emailQueue.grantConsumeMessages(emailLambda);

    // add sqs as event source
    emailLambda.addEventSource(new eventSources.SqsEventSource(emailQueue));
  }
}
