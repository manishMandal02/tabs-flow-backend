import { Construct } from 'constructs';

import { Stack, StackProps, aws_dynamodb } from 'aws-cdk-lib';

import { EmailService } from './email';

type ServiceStackProps = StackProps & {
  stage: string;
  appName: string;
  googleClientId: string;
  googleClientSecret: string;
  sesEmail: string;
  emailQueueName: string;
  database: aws_dynamodb.Table;
};

export class ServiceStack extends Stack {
  constructor(scope: Construct, id: string, props: ServiceStackProps) {
    super(scope, id, props);

    // TODO - give lambda execution role permissions and set AWS_REGION=ap-south-1

    // email service
    const emailService = new EmailService(this, 'EmailService', {
      stage: props.stage,
      ZEPTO_MAIL_API_KEY: process.env.ZEPTO_MAIL_API_KEY as string
    });

    // TODO - auth service
  }
}
