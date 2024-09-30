import { Construct } from 'constructs';

import {
  Stack,
  StackProps,
  aws_dynamodb,
  aws_apigateway as apiGateway,
  aws_iam as iam,
  Fn
} from 'aws-cdk-lib';

import { EmailService } from './email';
import { AuthService } from './auth';
import { config } from '../../../config';
import { UserService } from './users';
import { RestApi } from './rest-api';

type ServiceStackProps = StackProps & {
  stage: string;
};

export class ServiceStack extends Stack {
  constructor(scope: Construct, id: string, props: ServiceStackProps) {
    super(scope, id, props);

    const mainTableArn = Fn.importValue('MainTableArn');
    const sessionsTableArn = Fn.importValue('SessionsTableArn');

    const mainDB: aws_dynamodb.ITable = aws_dynamodb.Table.fromTableArn(this, 'MainTable', mainTableArn);
    const sessionsDB: aws_dynamodb.ITable = aws_dynamodb.Table.fromTableArn(
      this,
      'SessionsTable',
      sessionsTableArn
    );

    // create an IAM role for lambda
    const lambdaRole = new iam.Role(this, 'LambdaRole', {
      assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com')
    });

    // add basic execution role permissions
    lambdaRole.addManagedPolicy(
      iam.ManagedPolicy.fromAwsManagedPolicyName('service-role/AWSLambdaBasicExecutionRole')
    );

    const apiG = new RestApi(this, {
      stage: props.stage
    });

    const emailService = new EmailService(this, {
      lambdaRole,
      stage: props.stage
    });

    const authService = new AuthService(this, {
      lambdaRole,
      sessionsDB: sessionsDB,
      stage: props.stage,
      apiGW: apiG.restAPI,
      emailQueueURL: emailService.queueURL
    });

    new UserService(this, {
      stage: props.stage,
      apiGW: apiG.restAPI,
      lambdaRole,
      db: mainDB,
      apiAuthorizer: authService.apiAuthorizer,
      emailQueueURL: emailService.queueURL
    });
  }
}
