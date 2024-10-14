import { Construct } from 'constructs';

import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { aws_apigateway, aws_dynamodb, aws_iam, aws_sqs } from 'aws-cdk-lib';

import { config } from '../../../config';

type SpaceServiceProps = {
  stage: string;
  apiGW: aws_apigateway.RestApi;
  db: aws_dynamodb.ITable;
  lambdaRole: aws_iam.Role;
  apiAuthorizer: aws_apigateway.RequestAuthorizer;
  notificationQueue: aws_sqs.Queue;
};

export class SpaceService extends Construct {
  constructor(scope: Construct, props: SpaceServiceProps, id = 'SpaceService') {
    super(scope, id);

    const spaceServiceLambdaName = `${id}_${props.stage}`;
    const spaceServiceLambda = new GoFunction(this, spaceServiceLambdaName, {
      functionName: spaceServiceLambdaName,
      entry: '../cmd/spaces/main.go',
      runtime: config.lambda.Runtime,
      timeout: config.lambda.Timeout,
      memorySize: config.lambda.MemorySize,
      logRetention: config.lambda.LogRetention,
      role: props.lambdaRole,
      architecture: config.lambda.Architecture,
      bundling: config.lambda.GoBundling,
      environment: {
        DDB_MAIN_TABLE_NAME: props.db.tableName,
        NOTIFICATIONS_QUEUE_URL: props.notificationQueue.queueUrl
      }
    });

    // grant permissions to lambda to read/write to dynamodb and send message to queue
    props.db.grantReadWriteData(spaceServiceLambda);
    props.notificationQueue.grantSendMessages(spaceServiceLambda);

    // add spaces resource/endpoints to api gateway
    const spacesResource = props.apiGW.root.addResource('spaces').addProxy({ anyMethod: false });
    spacesResource.addMethod('ANY', new aws_apigateway.LambdaIntegration(spaceServiceLambda), {
      authorizationType: aws_apigateway.AuthorizationType.CUSTOM,
      authorizer: props.apiAuthorizer
    });
  }
}
