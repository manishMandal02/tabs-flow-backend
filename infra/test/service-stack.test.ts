import { stat } from 'fs';
import { PolicyDocument } from 'aws-cdk-lib/aws-iam';
import { EmailService } from './../lib/stacks/services/email';
import * as cdk from 'aws-cdk-lib';
import { Capture, Match, Template } from 'aws-cdk-lib/assertions';

import { config } from '../config';
import { ServiceStack } from './../lib/stacks/services/service-stack';

describe('ServiceStack', () => {
  const app = new cdk.App();
  const stage = config.Env.DEPLOY_STAGE;

  const serviceStack = new ServiceStack(app, 'ServiceStack', {
    stage,
    env: {
      region: process.env.AWS_REGION,
      account: process.env.AWS_ACCOUNT_ID
    },
    removalPolicy: stage === config.Stage.Prod ? cdk.RemovalPolicy.RETAIN : cdk.RemovalPolicy.DESTROY
  });
  const template = Template.fromStack(serviceStack);

  //   expect(template.toJSON()).toMatchSnapshot();

  const resources = template.findResources('AWS::Lambda::Function');

  //   console.log('🚀 email lambda: ', JSON.stringify(resources, null, 2));

  test('EmailService', () => {
    // assert lambda function
    template.hasResourceProperties('AWS::Lambda::Function', {
      FunctionName: `EmailService_${stage}`,
      Handler: 'bootstrap',
      Architectures: [config.Lambda.Architecture.name],
      Runtime: config.Lambda.Runtime.name,
      MemorySize: config.Lambda.MemorySize,
      Timeout: config.Lambda.Timeout.toSeconds(),
      Environment: {
        Variables: {
          ZEPTO_MAIL_API_KEY: stage === config.Stage.Test ? '' : expect.any(String),
          EMAIL_QUEUE_URL: {
            Ref: Match.stringLikeRegexp('EmailService')
          }
        }
      }
    });

    // assert sqs queue
    template.hasResourceProperties('AWS::SQS::Queue', {
      QueueName: `${config.AppName}-Emails_${stage}`,
      VisibilityTimeout: 300,
      DelaySeconds: 1,
      // assert dead letter queue
      RedrivePolicy: {
        deadLetterTargetArn: {
          'Fn::GetAtt': [Match.stringLikeRegexp('EmailService'), 'Arn']
        },
        maxReceiveCount: 3
      }
    });

    // assert lambda sqs event source
    template.hasResourceProperties('AWS::Lambda::EventSourceMapping', {
      FunctionName: {
        Ref: Match.stringLikeRegexp('EmailService')
      },
      BatchSize: 1,
      EventSourceArn: {
        'Fn::GetAtt': [Match.stringLikeRegexp('EmailService'), 'Arn']
      }
    });

    let iamStatement: any[] = [];

    let verifiedSQSIamPolicy = false;

    // verify lambda permission for sqs
    const policies = template.findResources('AWS::IAM::Policy', {});

    for (const policy of Object.values(policies)) {
      const PolicyDocument = policy['Properties']['PolicyDocument'] as any;

      if ((policy['Properties']['PolicyName'] as string).startsWith('LambdaRoleDefaultPolicy')) {
        const statement = PolicyDocument['Statement'] as any;
        const actions = statement[0]['Action'] as string[];
        if (Array.isArray(statement) && Array.isArray(actions) && actions[0].startsWith('sqs:')) {
          iamStatement = statement as any[];
          break;
        }
      }
    }

    expect(iamStatement.length).toBeGreaterThan(0);

    for (const statement of iamStatement) {
      if (
        statement.Effect !== 'Allow' ||
        !Array.isArray(statement.Action) ||
        !Array.isArray(statement.Resource)
      )
        continue;

      const actions = statement.Action as string[];

      console.log('🚀 ~ file: service-stack.test.ts:116 ~ test ~ actions:', actions);

      if (!actions[0].startsWith('sqs:')) continue;

      // iam policy for sqs permission
      expect(actions).toHaveLength(5);

      expect(actions).toContain('sqs:ReceiveMessage');
      expect(actions).toContain('sqs:DeleteMessage');

      const resources = statement.Resource as { [key: string]: string[] }[];

      console.log('🚀 ~ file: service-stack.test.ts:105 ~ test ~ resources:', resources);
      for (const resource of resources) {
        if (resource['Fn::GetAtt'][0].includes('EmailService') && resource['Fn::GetAtt'][1] === 'Arn') {
          verifiedSQSIamPolicy = true;
          break;
        }
      }
    }

    expect(verifiedSQSIamPolicy).toBeTruthy();
  });
});
