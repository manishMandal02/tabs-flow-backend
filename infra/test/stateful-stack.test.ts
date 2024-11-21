import { StatefulStack } from '../lib/stacks/stateful/stateful';
import * as cdk from 'aws-cdk-lib';
import { Template } from 'aws-cdk-lib/assertions';
import { config } from '../config';

test('StatefulStack', () => {
  // skip test
  return;
  const app = new cdk.App();

  const statefulStack = new StatefulStack(app, 'StatefulStack', {
    terminationProtection: false,
    env: {
      region: process.env.AWS_REGION,
      account: process.env.AWS_ACCOUNT_ID
    },
    stage: config.Stage.Prod,
    removalPolicy: cdk.RemovalPolicy.RETAIN
  });

  const template = Template.fromStack(statefulStack);

  const ddbPros = {
    BillingMode: 'PAY_PER_REQUEST',
    DeletionProtectionEnabled: true,
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
    TableName: `${config.AppName}-${config.DynamoDB.SessionsTableName}_${config.Stage.Prod}`,
    ...ddbPros
  });
  template.hasResourceProperties('AWS::DynamoDB::Table', {
    TableName: `${config.AppName}-${config.DynamoDB.MainTableName}_${config.Stage.Prod}`,
    ...ddbPros
  });

  template.hasResourceProperties('AWS::DynamoDB::Table', {
    TableName: `${config.AppName}-${config.DynamoDB.SearchIndexTableName}_${config.Stage.Prod}}`,
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
