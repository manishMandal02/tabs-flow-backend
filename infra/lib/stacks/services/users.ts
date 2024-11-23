import { Construct } from 'constructs';

import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { aws_apigateway, aws_dynamodb, aws_iam } from 'aws-cdk-lib';
import * as sqs from 'aws-cdk-lib/aws-sqs';

import { config } from '../../../config';

type UsersServiceProps = {
  stage: string;
  apiGW: aws_apigateway.RestApi;
  db: aws_dynamodb.ITable;
  lambdaRole: aws_iam.Role;
  emailQueue: sqs.Queue;
  apiAuthorizer: aws_apigateway.RequestAuthorizer;
};

export class UsersService extends Construct {
  constructor(scope: Construct, props: UsersServiceProps, id = 'UsersService') {
    super(scope, id);

    const userServiceLambdaName = `${id}_${props.stage}`;
    const usersServiceLambda = new GoFunction(this, userServiceLambdaName, {
      functionName: userServiceLambdaName,
      entry: '../cmd/users/main.go',
      runtime: config.Lambda.Runtime,
      timeout: config.Lambda.Timeout,
      memorySize: config.Lambda.MemorySize,
      logRetention: config.Lambda.LogRetention,
      role: props.lambdaRole,
      architecture: config.Lambda.Architecture,
      bundling: config.Lambda.GoBundling,
      environment: {
        EMAIL_QUEUE_URL: props.emailQueue.queueUrl,
        DDB_MAIN_TABLE_NAME: props.db.tableName
      }
    });

    // grant permissions to lambda to read/write to dynamodb and send message to email queue
    props.db.grantReadWriteData(usersServiceLambda);

    props.emailQueue.grantSendMessages(usersServiceLambda);

    // add users resource/endpoints to api gateway
    const usersResource = props.apiGW.root.addResource('users');

    usersResource.addMethod('ANY', new aws_apigateway.LambdaIntegration(usersServiceLambda), {
      authorizationType: aws_apigateway.AuthorizationType.CUSTOM,
      authorizer: props.apiAuthorizer
    });

    // add proxy resource
    const proxyResource = usersResource.addProxy({ anyMethod: false });

    // add method to proxy resource
    proxyResource.addMethod('ANY', new aws_apigateway.LambdaIntegration(usersServiceLambda), {
      authorizationType: aws_apigateway.AuthorizationType.CUSTOM,
      authorizer: props.apiAuthorizer
    });
  }
}
