import * as cdk from 'aws-cdk-lib';
import { Match, Template } from 'aws-cdk-lib/assertions';
import { describe, expect, test } from '@jest/globals';

import { config } from '../config';
import { ServiceStack } from '../lib/stacks/services/services';
import {
  assertLambdaAPIGatewayIntegration,
  assertLambdaFunction,
  assertLambdaSQSEventSourceMapping,
  assertSQSQueue,
  verifyLambdaSQSPermission
} from './helpers';

const serviceName = {
  Email: 'EmailService',
  Auth: 'AuthService',
  Users: 'UsersService',
  Notes: 'NotesService',
  Spaces: 'SpacesService',
  Notification: 'NotificationsService'
};

describe('ServiceStack', () => {
  const app = new cdk.App();
  const stage = config.Env.DEPLOY_STAGE;

  const serviceStack = new ServiceStack(app, 'ServiceStack', {
    stage,
    terminationProtection: stage === config.Stage.Prod,
    env: {
      region: process.env.AWS_REGION,
      account: process.env.AWS_ACCOUNT_ID
    },
    removalPolicy: stage === config.Stage.Prod ? cdk.RemovalPolicy.RETAIN : cdk.RemovalPolicy.DESTROY
  });
  const template = Template.fromStack(serviceStack);

  // email  service
  test('EmailService', () => {
    assertLambdaFunction({
      stage,
      template,
      service: serviceName.Email,
      env: {
        ZEPTO_MAIL_API_KEY: stage === config.Stage.Test ? '' : Match.anyValue(),
        EMAIL_QUEUE_URL: Match.anyValue()
      }
    });

    assertSQSQueue(template, `${config.AppName}-Emails_${stage}`, serviceName.Email);

    assertLambdaSQSEventSourceMapping(template, serviceName.Email);

    const verifiedSQSIamPolicy = verifyLambdaSQSPermission(template, serviceName.Email);

    expect(verifiedSQSIamPolicy).toBeTruthy();
  });

  // auth service
  test('AuthService', () => {
    // auth service lambda
    assertLambdaFunction({
      stage,
      template,
      service: serviceName.Auth,
      env: {
        JWT_SECRET_KEY: Match.stringLikeRegexp(config.AppName),
        EMAIL_QUEUE_URL: Match.anyValue(),
        DDB_SESSIONS_TABLE_NAME: Match.anyValue()
      }
    });

    // authorizer lambda
    assertLambdaFunction({
      stage,
      template,
      service: serviceName.Auth,
      name: 'Authorizer_' + stage,
      env: {
        JWT_SECRET_KEY: Match.stringLikeRegexp(config.AppName),
        DDB_SESSIONS_TABLE_NAME: Match.anyValue()
      }
    });

    assertLambdaAPIGatewayIntegration({ template, service: serviceName.Auth, baseURL: 'auth' });
  });

  test('UsersService', () => {
    assertLambdaFunction({
      stage,
      template,
      service: serviceName.Users,
      env: {
        EMAIL_QUEUE_URL: Match.anyValue(),
        DDB_MAIN_TABLE_NAME: Match.anyValue()
      }
    });

    assertLambdaAPIGatewayIntegration({
      template,
      service: serviceName.Users,
      baseURL: 'users',
      hasAuthorization: true
    });
  });

  test('SpacesService', () => {
    assertLambdaFunction({
      stage,
      template,
      service: serviceName.Spaces,
      env: {
        DDB_MAIN_TABLE_NAME: Match.anyValue(),
        NOTIFICATIONS_QUEUE_URL: Match.anyValue()
      }
    });

    assertLambdaAPIGatewayIntegration({
      template,
      service: serviceName.Spaces,
      baseURL: 'spaces',
      hasAuthorization: true
    });
  });

  test('NotesService', () => {
    assertLambdaFunction({
      stage,
      template,
      service: serviceName.Notes,
      env: {
        DDB_MAIN_TABLE_NAME: Match.anyValue(),
        DDB_SEARCH_INDEX_TABLE_NAME: Match.anyValue(),
        NOTIFICATIONS_QUEUE_URL: Match.anyValue()
      }
    });

    assertLambdaAPIGatewayIntegration({
      template,
      service: serviceName.Notes,
      baseURL: 'notes',
      hasAuthorization: true
    });
  });

  test('NotificationsService', () => {
    assertLambdaFunction({
      stage,
      template,
      service: serviceName.Notification,
      env: {
        DDB_MAIN_TABLE_NAME: Match.anyValue(),
        NOTIFICATIONS_QUEUE_ARN: Match.anyValue(),
        SCHEDULER_ROLE_ARN: Match.anyValue(),
        NOTIFICATIONS_QUEUE_URL: Match.anyValue(),
        VAPID_PRIVATE_KEY: Match.anyValue(),
        VAPID_PUBLIC_KEY: Match.anyValue()
      }
    });

    assertSQSQueue(template, `${config.AppName}-Notifications_${stage}`, serviceName.Notification);

    assertLambdaSQSEventSourceMapping(template, serviceName.Notification);

    // assert sqs permission
    template.hasResourceProperties('AWS::IAM::Policy', {
      PolicyDocument: {
        Statement: [
          {
            Action: Match.arrayWith(['sqs:SendMessage', 'sqs:GetQueueAttributes', 'sqs:GetQueueUrl']),
            Effect: 'Allow',
            Resource: {
              'Fn::GetAtt': [Match.stringLikeRegexp(serviceName.Notification), 'Arn']
            }
          }
        ],
        Version: '2012-10-17'
      }
    });

    // assert scheduler role
    template.hasResourceProperties('AWS::IAM::Role', {
      AssumeRolePolicyDocument: {
        Statement: [
          {
            Action: 'sts:AssumeRole',
            Effect: 'Allow',
            Principal: {
              Service: 'scheduler.amazonaws.com'
            }
          }
        ],
        Version: '2012-10-17'
      },
      Description: Match.stringLikeRegexp('EventBridge Scheduler')
    });

    assertLambdaAPIGatewayIntegration({
      template,
      service: serviceName.Notification,
      baseURL: 'notifications',
      hasAuthorization: true
    });
  });
});
