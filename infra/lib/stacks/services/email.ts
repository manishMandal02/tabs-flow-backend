import { Construct } from 'constructs';

import { Duration, aws_iam } from 'aws-cdk-lib';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import * as eventSources from 'aws-cdk-lib/aws-lambda-event-sources';

import { config } from '../../../config';

type EmailServiceProps = {
  stage: string;
  lambdaRole: aws_iam.Role;
};

export class EmailService extends Construct {
  Queue: sqs.Queue;
  constructor(scope: Construct, props: EmailServiceProps, id = 'EmailService') {
    super(scope, id);

    const { ZEPTO_MAIL_API_KEY } = config.Env;

    const queueName = `${config.AppName}-Emails_${props.stage}`;
    // sqs queue
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

    const emailServiceLambdaName = id + '_' + props.stage;

    const emailServiceFunction = new GoFunction(this, emailServiceLambdaName, {
      functionName: emailServiceLambdaName,
      entry: '../cmd/email/main.go',
      runtime: config.lambda.Runtime,
      timeout: config.lambda.Timeout,
      memorySize: config.lambda.MemorySize,
      logRetention: config.lambda.LogRetention,
      role: props.lambdaRole,
      architecture: config.lambda.Architecture,
      bundling: config.lambda.GoBundling,
      environment: {
        ZEPTO_MAIL_API_KEY,
        EMAIL_QUEUE_URL: emailQueue.queueUrl
      }
    });

    // grants permissions to lambda
    emailQueue.grantConsumeMessages(emailServiceFunction);

    // add sqs as event source
    emailServiceFunction.addEventSource(new eventSources.SqsEventSource(emailQueue, { batchSize: 1 }));

    this.Queue = emailQueue;
  }
}
