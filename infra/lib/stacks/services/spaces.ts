import { Construct } from 'constructs';

import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { aws_apigateway, aws_dynamodb, aws_iam, aws_sqs } from 'aws-cdk-lib';

import { config } from '../../../config';

type SpacesServiceProps = {
  stage: string;
  apiGW: aws_apigateway.RestApi;
  db: aws_dynamodb.ITable;
  lambdaRole: aws_iam.Role;
  apiAuthorizer: aws_apigateway.RequestAuthorizer;
  notificationQueue: aws_sqs.Queue;
};

export class SpacesService extends Construct {
  constructor(scope: Construct, props: SpacesServiceProps, id = 'SpacesService') {
    super(scope, id);

    const spaceServiceLambdaName = `${id}_${props.stage}`;
    const spaceServiceLambda = new GoFunction(this, spaceServiceLambdaName, {
      functionName: spaceServiceLambdaName,
      entry: '../cmd/spaces/main.go',
      runtime: config.Lambda.Runtime,
      timeout: config.Lambda.Timeout,
      memorySize: config.Lambda.MemorySize,
      logRetention: config.Lambda.LogRetention,
      role: props.lambdaRole,
      architecture: config.Lambda.Architecture,
      bundling: config.Lambda.GoBundling,
      environment: {
        DDB_MAIN_TABLE_NAME: props.db.tableName,
        NOTIFICATIONS_QUEUE_URL: props.notificationQueue.queueUrl
      }
    });

    // grant permissions to lambda to read/write to dynamodb and send message to queue
    props.db.grantReadWriteData(spaceServiceLambda);
    props.notificationQueue.grantSendMessages(spaceServiceLambda);

    // add spaces resource/endpoints to api gateway
    const spacesResource = props.apiGW.root.addResource('spaces');

    spacesResource.addMethod('ANY', new aws_apigateway.LambdaIntegration(spaceServiceLambda), {
      authorizationType: aws_apigateway.AuthorizationType.CUSTOM,
      authorizer: props.apiAuthorizer
    });

    // add proxy resource
    const proxyResource = spacesResource.addProxy({ anyMethod: false });

    // add method to proxy resource
    proxyResource.addMethod('ANY', new aws_apigateway.LambdaIntegration(spaceServiceLambda), {
      authorizationType: aws_apigateway.AuthorizationType.CUSTOM,
      authorizer: props.apiAuthorizer
    });
  }
}
