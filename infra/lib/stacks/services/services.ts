import { Construct } from 'constructs';
import {
  Stack,
  StackProps,
  aws_dynamodb,
  aws_iam as iam,
  aws_ssm as ssm,
  Lazy,
  RemovalPolicy
} from 'aws-cdk-lib';

import { RestApi } from './rest-api';
import { AuthService } from './auth';
import { UsersService } from './users';
import { EmailService } from './email';
import { NotesService } from './notes';
import { SpacesService } from './spaces';
import { NotificationsService } from './notifications';
import { config } from '../../../config';

type ServiceStackProps = StackProps & {
  stage: string;
  removalPolicy: RemovalPolicy;
};

export class ServiceStack extends Stack {
  constructor(scope: Construct, id: string, props: ServiceStackProps) {
    super(scope, id, props);

    const mainTableArn = Lazy.string({
      produce: () => ssm.StringParameter.valueForStringParameter(this, config.SSMParameterName.MainTableArn)
    });

    const sessionsTableArn = Lazy.string({
      produce: () =>
        ssm.StringParameter.valueForStringParameter(this, config.SSMParameterName.SessionsTableArn)
    });
    const searchIndexTableArn = Lazy.string({
      produce: () =>
        ssm.StringParameter.valueForStringParameter(this, config.SSMParameterName.SearchIndexTableArn)
    });

    const apiDomainCertArn = Lazy.string({
      produce: () =>
        ssm.StringParameter.valueForStringParameter(this, config.SSMParameterName.APIDomainCertArn)
    });

    if (!mainTableArn || !sessionsTableArn || !searchIndexTableArn) {
      throw new Error('Missing required DynamoDB Table ARNs SSM parameters.');
    }

    if (props.stage !== config.Stage.Test && !apiDomainCertArn) {
      throw new Error('Missing required API Domain Certificate ARN SSM parameter.');
    }

    const mainDB: aws_dynamodb.ITable = aws_dynamodb.Table.fromTableArn(this, 'MainTableAr', mainTableArn);

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
      stage: props.stage,
      domainCertArn: apiDomainCertArn
    });

    const emailService = new EmailService(this, {
      lambdaRole,
      stage: props.stage,
      removalPolicy: props.removalPolicy
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
      apiAuthorizer: authService.apiAuthorizer,
      removalPolicy: props.removalPolicy
    });

    new UsersService(this, {
      lambdaRole,
      db: mainDB,
      stage: props.stage,
      apiGW: apiG.restAPI,
      apiAuthorizer: authService.apiAuthorizer,
      emailQueue: emailService.Queue
    });

    new SpacesService(this, {
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
