import { Construct } from 'constructs';

import { aws_lambda, aws_apigateway as apiGateway, Duration, aws_dynamodb, aws_iam } from 'aws-cdk-lib';
import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { config } from '../../../config';

type AuthServiceProps = {
  stage: string;
  apiGW: apiGateway.RestApi;
  sessionDB: aws_dynamodb.Table;
  emailQueueURL: string;
  lambdaRole: aws_iam.Role;
};

export class AuthService extends Construct {
  apiAuthorizer: apiGateway.RequestAuthorizer;
  constructor(scope: Construct, props: AuthServiceProps, id: string = 'AuthService') {
    super(scope, id);

    const { JWT_SECRET_KEY } = config.Env;

    const authServiceLambda = new GoFunction(this, `${id}-${props.stage}`, {
      entry: '../cmd/auth/main.go',
      runtime: aws_lambda.Runtime.PROVIDED_AL2,
      timeout: config.lambda.Timeout,
      memorySize: config.lambda.MemorySize,
      logRetention: config.lambda.LogRetention,
      role: props.lambdaRole,
      bundling: config.lambda.GoBundling,
      environment: {
        JWT_SECRET_KEY,
        EMAIL_SQS_QUEUE_URL: props.emailQueueURL,
        DDB_SESSIONS_TABLE_NAME: props.sessionDB.tableName
      }
    });

    props.sessionDB.grantReadWriteData(authServiceLambda);

    // add auth resource/endpoints to api gateway
    const authResource = props.apiGW.root.addResource('auth');

    authResource.addMethod('ANY', new apiGateway.LambdaIntegration(authServiceLambda));

    //*-- authorizer --
    const authorizerLambda = new GoFunction(this, `Authorizer-${props.stage}`, {
      entry: '../cmd/auth/lambda_authorizer/main.go',
      runtime: aws_lambda.Runtime.PROVIDED_AL2,
      timeout: config.lambda.Timeout,
      memorySize: config.lambda.MemorySize,
      logRetention: config.lambda.LogRetention,
      role: props.lambdaRole,
      bundling: config.lambda.GoBundling,
      environment: {
        JWT_SECRET_KEY,
        DDB_SESSIONS_TABLE_NAME: props.sessionDB.tableName
      }
    });

    const authorizerName = config.AppName + '-Authorizer';

    // Lambda Authorizer with 'REQUEST' type
    const authorizer = new apiGateway.RequestAuthorizer(this, authorizerName, {
      authorizerName,
      handler: authorizerLambda,
      identitySources: [apiGateway.IdentitySource.header('cookies')],
      resultsCacheTtl: Duration.minutes(5)
    });

    props.sessionDB.grantReadWriteData(authorizerLambda);

    this.apiAuthorizer = authorizer;
  }
}
