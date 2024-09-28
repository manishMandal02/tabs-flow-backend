import { Construct } from 'constructs';
import { CfnOutput, RemovalPolicy, Stack, StackProps, aws_dynamodb } from 'aws-cdk-lib';
import { config } from '../../../config';

type StatefulStackProps = StackProps & {
  stage: string;
};

export class StatefulStack extends Stack {
  // go lambda triggers

  constructor(scope: Construct, id: string, props: StatefulStackProps) {
    super(scope, id, props);

    const mainTableName = `${config.AppName}-Main_${props.stage}`;
    const sessionsTableName = `${config.AppName}-Sessions_${props.stage}`;

    const mainTable = new aws_dynamodb.Table(this, mainTableName, {
      tableName: mainTableName,
      billingMode: aws_dynamodb.BillingMode.PAY_PER_REQUEST,
      partitionKey: {
        name: config.dynamoDB.PrimaryKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      sortKey: {
        name: config.dynamoDB.SortKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      removalPolicy: props.stage === config.Dev.Stage ? RemovalPolicy.DESTROY : RemovalPolicy.RETAIN
    });

    const sessionsTable = new aws_dynamodb.Table(this, sessionsTableName, {
      tableName: sessionsTableName,
      billingMode: aws_dynamodb.BillingMode.PAY_PER_REQUEST,
      partitionKey: {
        name: config.dynamoDB.PrimaryKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      sortKey: {
        name: config.dynamoDB.SortKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      timeToLiveAttribute: config.dynamoDB.TTL,
      removalPolicy: props.stage === config.Dev.Stage ? RemovalPolicy.DESTROY : RemovalPolicy.RETAIN
    });

    new CfnOutput(this, 'MainTableArn', {
      exportName: 'MainTableArn',
      value: mainTable.tableArn
    });

    new CfnOutput(this, 'SessionsTableArn', {
      exportName: 'SessionsTableArn',
      value: sessionsTable.tableArn
    });
  }
}
