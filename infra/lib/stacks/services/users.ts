import { Construct } from 'constructs';

import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { aws_apigateway, aws_dynamodb, aws_iam, aws_lambda } from 'aws-cdk-lib';

import { config } from '../../../config';
import { RetentionDays } from 'aws-cdk-lib/aws-logs';

type UserServiceProps = {
  stage: string;
  apiGW: aws_apigateway.RestApi;
  db: aws_dynamodb.Table;
  lambdaRole: aws_iam.Role;
  emailQueueURL: string;
};

export class UserService extends Construct {
  constructor(scope: Construct, id: string, props: UserServiceProps) {
    super(scope, id);

    const usersServiceLambda = new GoFunction(this, `${id}-${props.stage}`, {
      entry: 'cmd/users/main.go',
      runtime: aws_lambda.Runtime.PROVIDED_AL2,
      timeout: config.lambda.Timeout,
      memorySize: config.lambda.MemorySize,
      logRetention: config.lambda.LogRetention,
      role: props.lambdaRole,
      environment: {
        EMAIL_SQS_QUEUE_URL: props.emailQueueURL,
        DDB_SESSIONS_TABLE_NAME: props.db.tableName
      }
    });

    // grant permissions to lambda to read/write from dynamoDB
    props.db.grantReadWriteData(usersServiceLambda);

    // add users resource/endpoints to api gateway
    const authResource = props.apiGW.root.addResource('users');
    authResource.addMethod('ANY', new aws_apigateway.LambdaIntegration(usersServiceLambda));
  }
}
