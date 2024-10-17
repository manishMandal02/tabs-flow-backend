import { Construct } from 'constructs';

import { Stack, StackProps, aws_dynamodb, aws_iam as iam, aws_ssm as ssm, Fn, Lazy } from 'aws-cdk-lib';

import { EmailService } from './email';
import { AuthService } from './auth';
import { UserService } from './users';
import { RestApi } from './rest-api';
import { SpaceService } from './spaces';
import { NotesService } from './notes';
import { NotificationsService } from './notifications';

type ServiceStackProps = StackProps & {
  stage: string;
};

export class ServiceStack extends Stack {
  constructor(scope: Construct, id: string, props: ServiceStackProps) {
    super(scope, id, props);

    const mainTableArn = Lazy.string({
      produce: () => ssm.StringParameter.valueFromLookup(this, '/main-table-arn')
    });
    const sessionsTableArn = Lazy.string({
      produce: () => ssm.StringParameter.valueFromLookup(this, '/sessions-table-arn')
    });
    const searchIndexTableArn = Lazy.string({
      produce: () => ssm.StringParameter.valueFromLookup(this, '/search-index-table-arn')
    });

    const mainDB: aws_dynamodb.ITable = aws_dynamodb.Table.fromTableArn(this, 'MainTable', mainTableArn);
    const sessionsDB: aws_dynamodb.ITable = aws_dynamodb.Table.fromTableArn(
      this,
      'SessionsTable',
      sessionsTableArn
    );

    const searchIndexDB = aws_dynamodb.Table.fromTableArn(this, 'SearchIndexTable', searchIndexTableArn);

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
      sessionsDB,
      stage: props.stage,
      apiGW: apiG.restAPI,
      emailQueue: emailService.Queue
    });

    const notificationsService = new NotificationsService(this, {
      lambdaRole,
      db: mainDB,
      stage: props.stage,
      apiGW: apiG.restAPI,
      emailQueue: emailService.Queue,
      apiAuthorizer: authService.apiAuthorizer
    });

    new UserService(this, {
      lambdaRole,
      db: mainDB,
      stage: props.stage,
      apiGW: apiG.restAPI,
      apiAuthorizer: authService.apiAuthorizer,
      emailQueue: emailService.Queue
    });

    new SpaceService(this, {
      lambdaRole,
      db: mainDB,
      stage: props.stage,
      apiGW: apiG.restAPI,
      apiAuthorizer: authService.apiAuthorizer,
      notificationQueue: notificationsService.Queue
    });

    new NotesService(this, {
      searchIndexDB,
      mainDB,
      lambdaRole,
      apiGW: apiG.restAPI,
      stage: props.stage,
      apiAuthorizer: authService.apiAuthorizer,
      notificationQueue: notificationsService.Queue
    });
  }
}
