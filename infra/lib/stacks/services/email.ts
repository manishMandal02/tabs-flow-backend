import { Construct } from 'constructs';

import { Duration, aws_lambda } from 'aws-cdk-lib';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import * as eventSources from 'aws-cdk-lib/aws-lambda-event-sources';

import { config } from '../../../config';

type EmailServiceProps = {
  stage: string;
  ZEPTO_MAIL_API_KEY: string;
};

export class EmailService extends Construct {
  constructor(scope: Construct, id: string, props: EmailServiceProps) {
    super(scope, id);

    const queueName = `${config.AppName.toLowerCase()}_emailsv-${props.stage}`;
    //- sqs queue
    const dlqEmail = new sqs.Queue(this, queueName + '-dlq', {
      visibilityTimeout: Duration.seconds(300)
    });

    const emailQueue = new sqs.Queue(this, queueName, {
      queueName: queueName,
      visibilityTimeout: Duration.seconds(300),
      deliveryDelay: Duration.seconds(1),
      deadLetterQueue: {
        queue: dlqEmail,
        maxReceiveCount: 3
      }
    });

    const emailServiceFunction = new GoFunction(this, `${id}-${props.stage}`, {
      entry: 'cmd/email/main.go',
      runtime: aws_lambda.Runtime.PROVIDED_AL2,
      timeout: Duration.seconds(30),
      memorySize: 128,
      environment: {
        ZEPTO_MAIL_API_KEY: props.ZEPTO_MAIL_API_KEY
      }
    });

    // grants permissions to lambda
    emailQueue.grantConsumeMessages(emailServiceFunction);

    // add sqs as event source
    emailServiceFunction.addEventSource(new eventSources.SqsEventSource(emailQueue));
  }
}
