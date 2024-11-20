import { StatefulStack } from './../lib/stacks/stateful/stateful';
import * as cdk from 'aws-cdk-lib';
import { Template } from 'aws-cdk-lib/assertions';
import { config } from '../config';

// example test. To run these tests, uncomment this file along with the
// example resource in lib/infrastructure-stack.ts
test('StatefulStack ', () => {
  const app = new cdk.App();

  const statefulStack = new StatefulStack(app, 'StatefulStack', {
    terminationProtection: false,
    env: {
      region: process.env.AWS_REGION,
      account: process.env.AWS_ACCOUNT_ID
    },
    stage: config.Env.DEPLOY_STAGE,
    removalPolicy: cdk.RemovalPolicy.DESTROY
  });

  const template = Template.fromStack(statefulStack);

  template.hasResourceProperties('AWS::SQS::Queue', {
    VisibilityTimeout: 300
  });
});
