import { Construct } from 'constructs';
import { RemovalPolicy, Stack, StackProps, aws_dynamodb } from 'aws-cdk-lib';
import { config } from '../../../config';

type StatefulStackProps = StackProps & {
  stage: string;
  appName: string;
  googleClientId: string;
  googleClientSecret: string;
  sesEmail: string;
  emailQueueName: string;
};

export class StatefulStack extends Stack {
  // go lambda triggers

  database: aws_dynamodb.Table;

  constructor(scope: Construct, id: string, props: StatefulStackProps) {
    super(scope, id, props);

    const tableName = `${config.DDB.MainTableName} ${props.stage}`;

    const table = new aws_dynamodb.Table(this, tableName, {
      tableName: tableName,
      billingMode: aws_dynamodb.BillingMode.PAY_PER_REQUEST,
      partitionKey: {
        name: config.DDB.PrimaryKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      sortKey: {
        name: config.DDB.SortKey,
        type: aws_dynamodb.AttributeType.STRING
      },
      timeToLiveAttribute: config.DDB.TTL,
      removalPolicy: props.stage === config.Dev.Stage ? RemovalPolicy.DESTROY : RemovalPolicy.RETAIN
    });

    this.database = table;
  }
}
