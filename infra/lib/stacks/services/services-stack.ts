import { Construct } from 'constructs';

import { Stack, StackProps, aws_dynamodb, aws_apigateway as apiGateway, aws_iam as iam } from 'aws-cdk-lib';

import { EmailService } from './email';
import { AuthService } from './auth';
import { config } from '../../../config';
import { UserService } from './users';

type ServiceStackProps = StackProps & {
  stage: string;
  mainDB: aws_dynamodb.Table;
  sessionsDB: aws_dynamodb.Table;
};

export class ServiceStack extends Stack {
  constructor(scope: Construct, id: string, props: ServiceStackProps) {
    super(scope, id, props);

    // create an IAM role for lambda
    const lambdaRole = new iam.Role(this, 'LambdaRole', {
      assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com')
    });

    // add basic execution role permissions
    lambdaRole.addManagedPolicy(iam.ManagedPolicy.fromAwsManagedPolicyName('AWSLambdaBasicExecutionRole'));

    const resAPI = new apiGateway.RestApi(this, `${config.AppName}-${props.stage}`);

    const emailService = new EmailService(this, 'EmailService', {
      lambdaRole,
      stage: props.stage
    });

    new AuthService(this, 'AuthService', {
      lambdaRole,
      stage: props.stage,
      apiGW: resAPI,
      sessionDB: props.sessionsDB,
      emailQueueURL: emailService.queueURL
    });

    new UserService(this, 'UserService', {
      stage: props.stage,
      apiGW: resAPI,
      lambdaRole,
      db: props.mainDB,
      emailQueueURL: emailService.queueURL
    });
  }
}
