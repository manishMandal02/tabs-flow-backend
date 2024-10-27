import { Construct } from 'constructs';
import { CfnOutput, RemovalPolicy, Stack, StackProps, aws_dynamodb, aws_ssm as ssm } from 'aws-cdk-lib';
import { config } from '../../../config';

type StatefulStackProps = StackProps & {
  stage: string;
  removalPolicy: RemovalPolicy;
};

export class StatefulStack extends Stack {
  // go lambda triggers

  constructor(scope: Construct, id: string, props: StatefulStackProps) {
    super(scope, id, props);

    const mainTableName = `${config.AppName}-${config.DynamoDB.MainTableName}_${props.stage}`;
    const sessionsTableName = `${config.AppName}-${config.DynamoDB.SessionsTableName}_${props.stage}`;
    const searchIndexTableName = `${config.AppName}-${config.DynamoDB.SearchIndexTableName}_${props.stage}`;

    const mainTable = new aws_dynamodb.Table(this, mainTableName, {
      tableName: mainTableName,
      billingMode: aws_dynamodb.BillingMode.PAY_PER_REQUEST,
      partitionKey: {
        name: config.DynamoDB.PrimaryKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      sortKey: {
        name: config.DynamoDB.SortKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      removalPolicy: props.removalPolicy
    });

    const sessionsTable = new aws_dynamodb.Table(this, sessionsTableName, {
      tableName: sessionsTableName,
      billingMode: aws_dynamodb.BillingMode.PAY_PER_REQUEST,
      partitionKey: {
        name: config.DynamoDB.PrimaryKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      sortKey: {
        name: config.DynamoDB.SortKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      timeToLiveAttribute: config.DynamoDB.TTL,
      removalPolicy: props.removalPolicy
    });

    const searchIndexTable = new aws_dynamodb.Table(this, searchIndexTableName, {
      tableName: searchIndexTableName,
      billingMode: aws_dynamodb.BillingMode.PAY_PER_REQUEST,
      partitionKey: {
        name: config.DynamoDB.PrimaryKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      sortKey: {
        name: config.DynamoDB.SortKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      removalPolicy: props.removalPolicy
    });

    new ssm.StringParameter(this, 'MainTableArn', {
      parameterName: `/main-table-arn`,
      stringValue: mainTable.tableArn,
      tier: ssm.ParameterTier.STANDARD
    });
    new ssm.StringParameter(this, 'SessionsTableArn', {
      parameterName: `/sessions-table-arn`,
      stringValue: sessionsTable.tableArn,
      tier: ssm.ParameterTier.STANDARD
    });
    new ssm.StringParameter(this, 'SearchIndexTableArn', {
      parameterName: `/search-index-table-arn`,
      stringValue: searchIndexTable.tableArn,
      tier: ssm.ParameterTier.STANDARD
    });
  }
}
