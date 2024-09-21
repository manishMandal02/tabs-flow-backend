import { Construct } from 'constructs';

import { Stack, StackProps, aws_dynamodb, aws_apigateway as apiGateway } from 'aws-cdk-lib';

import { EmailService } from './email';
import { AuthService } from './auth';
import { config } from '../../../config';

type ServiceStackProps = StackProps & {
  stage: string;
  mainDB: aws_dynamodb.Table;
  sessionsDB: aws_dynamodb.Table;
};

export class ServiceStack extends Stack {
  constructor(scope: Construct, id: string, props: ServiceStackProps) {
    super(scope, id, props);

    // TODO - give lambda execution role permissions and set AWS_REGION=ap-south-1

    const resAPI = new apiGateway.RestApi(this, `${id}-${props.stage}`);

    // email service
    const emailService = new EmailService(this, 'EmailService', {
      stage: props.stage
    });

    const authService = new AuthService(this, 'AuthService', {
      stage: props.stage,
      apiGW: resAPI,
      sessionDB: props.sessionsDB,
      emailQueueURL: emailService.queueURL
    });

    // TODO - user service
  }
}
