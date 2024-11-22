import { StatefulStack } from '../lib/stacks/stateful/stateful';
import * as cdk from 'aws-cdk-lib';
import { Template } from 'aws-cdk-lib/assertions';
import { config } from '../config';

test('StatefulStack', () => {
  const stage = config.Env.DEPLOY_STAGE;

  const app = new cdk.App();

  const statefulStack = new StatefulStack(app, 'StatefulStack', {
    stage,
    terminationProtection: stage === config.Stage.Prod,
    env: {
      region: process.env.AWS_REGION,
      account: process.env.AWS_ACCOUNT_ID
    },
    removalPolicy: stage === config.Stage.Prod ? cdk.RemovalPolicy.RETAIN : cdk.RemovalPolicy.DESTROY
  });

  const template = Template.fromStack(statefulStack);

  const ddbPros = {
    BillingMode: 'PAY_PER_REQUEST',
    DeletionProtectionEnabled: stage === config.Stage.Prod,
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
    TableName: `${config.DynamoDB.SessionsTableName}`,
    ...ddbPros
  });
  template.hasResourceProperties('AWS::DynamoDB::Table', {
    TableName: `${config.DynamoDB.MainTableName}`,
    ...ddbPros
  });

  template.hasResourceProperties('AWS::DynamoDB::Table', {
    TableName: `${config.DynamoDB.SearchIndexTableName}`,
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
