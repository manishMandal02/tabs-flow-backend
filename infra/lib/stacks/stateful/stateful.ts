import { Construct } from 'constructs';
import { RemovalPolicy, Stack, StackProps, aws_dynamodb, aws_ssm as ssm } from 'aws-cdk-lib';
import { config } from '../../../config';

type StatefulStackProps = StackProps & {
  stage: string;
  removalPolicy: RemovalPolicy;
};

export class StatefulStack extends Stack {
  // go lambda triggers

  constructor(scope: Construct, id: string, props: StatefulStackProps) {
    super(scope, id, props);

    const commonTableProps = {
      billingMode: aws_dynamodb.BillingMode.PAY_PER_REQUEST,
      partitionKey: {
        name: config.DynamoDB.PrimaryKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      sortKey: {
        name: config.DynamoDB.SortKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      removalPolicy: props.removalPolicy,
      deletionProtection: props.removalPolicy === RemovalPolicy.RETAIN
    };

    const mainTable = new aws_dynamodb.Table(this, config.DynamoDB.MainTableName, {
      tableName: config.DynamoDB.MainTableName,
      pointInTimeRecovery: props.removalPolicy === RemovalPolicy.RETAIN,
      ...commonTableProps
    });

    const searchIndexTable = new aws_dynamodb.Table(this, config.DynamoDB.SearchIndexTableName, {
      tableName: config.DynamoDB.SearchIndexTableName,
      pointInTimeRecovery: props.removalPolicy === RemovalPolicy.RETAIN,
      ...commonTableProps
    });

    const sessionsTable = new aws_dynamodb.Table(this, config.DynamoDB.SessionsTableName, {
      tableName: config.DynamoDB.SessionsTableName,
      timeToLiveAttribute: config.DynamoDB.TTL,
      ...commonTableProps
    });

    // save table arns in ssm parameters store
    new ssm.StringParameter(this, 'MainTableArn', {
      parameterName: config.SSMParameterName.MainTableArn,
      stringValue: mainTable.tableArn,
      tier: ssm.ParameterTier.STANDARD
    });

    new ssm.StringParameter(this, 'SearchIndexTableArn', {
      parameterName: config.SSMParameterName.SearchIndexTableArn,
      stringValue: searchIndexTable.tableArn,
      tier: ssm.ParameterTier.STANDARD
    });

    new ssm.StringParameter(this, 'SessionsTableArn', {
      parameterName: config.SSMParameterName.SessionsTableArn,
      stringValue: sessionsTable.tableArn,
      tier: ssm.ParameterTier.STANDARD
    });
  }
}
