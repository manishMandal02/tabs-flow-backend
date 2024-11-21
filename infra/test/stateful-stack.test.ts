import { StatefulStack } from '../lib/stacks/stateful/stateful';
import * as cdk from 'aws-cdk-lib';
import { Template } from 'aws-cdk-lib/assertions';
import { config } from '../config';
import { PartitionKey } from 'aws-cdk-lib/aws-appsync';
import { Key } from 'aws-cdk-lib/aws-kms';

// example test. To run these tests, uncomment this file along with the
// example resource in lib/infrastructure-stack.ts
test('StatefulStack has 3 DynamoDB tables', () => {
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

  const ddbPros = {
    BillingMode: 'PAY_PER_REQUEST',
    DeletionProtectionEnabled: config.Stage.Test !== config.Env.DEPLOY_STAGE,
    KeySchema: [
      {
        AttributeName: config.DynamoDB.PrimaryKey,
        KeyType: 'HASH'
      },
      {
        AttributeName: config.DynamoDB.SortKey,
        KeyType: 'RANGE'
      }
    ]
  };

  // assert DynamoDB tables
  template.hasResourceProperties('AWS::DynamoDB::Table', {
    TableName: `${config.AppName}-${config.DynamoDB.SessionsTableName}_${config.Env.DEPLOY_STAGE}`,
    ...ddbPros
  });
  template.hasResourceProperties('AWS::DynamoDB::Table', {
    TableName: `${config.AppName}-${config.DynamoDB.MainTableName}_${config.Env.DEPLOY_STAGE}`,
    ...ddbPros
  });

  template.hasResourceProperties('AWS::DynamoDB::Table', {
    TableName: `${config.AppName}-${config.DynamoDB.SearchIndexTableName}_${config.Env.DEPLOY_STAGE}`,
    ...ddbPros
  });

  // assert SSM parameters
  template.hasResourceProperties('AWS::SSM::Parameter', {
    Name: config.SSMParameterName.SessionsTableArn,
    Type: 'String'
  });
  template.hasResourceProperties('AWS::SSM::Parameter', {
    Name: config.SSMParameterName.MainTableArn,
    Type: 'String'
  });
  template.hasResourceProperties('AWS::SSM::Parameter', {
    Name: config.SSMParameterName.SearchIndexTableArn,
    Type: 'String'
  });
});
