import { Construct } from 'constructs';

import { aws_lambda, aws_apigateway as apiGateway, Duration, aws_dynamodb, aws_iam } from 'aws-cdk-lib';
import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { config } from '../../../config';

type AuthServiceProps = {
  stage: string;
  apiGW: apiGateway.RestApi;
  sessionsDB: aws_dynamodb.ITable;
  emailQueueURL: string;
  lambdaRole: aws_iam.Role;
};

export class AuthService extends Construct {
  apiAuthorizer: apiGateway.RequestAuthorizer;
  constructor(scope: Construct, props: AuthServiceProps, id: string = 'AuthService') {
    super(scope, id);

    const { JWT_SECRET_KEY } = config.Env;

    const authLambdaName = id + '_' + props.stage;

    const authServiceLambda = new GoFunction(this, authLambdaName, {
      functionName: authLambdaName,
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
        DDB_SESSIONS_TABLE_NAME: props.sessionsDB.tableName
      }
    });

    props.sessionsDB.grantReadWriteData(authServiceLambda);

    // add auth resource/endpoints to api gateway
    const authResource = props.apiGW.root.addResource('auth');

    authResource.addMethod('ANY', new apiGateway.LambdaIntegration(authServiceLambda));

    //*-- authorizer --
    const authorizerLambdaName = 'Authorizer_' + props.stage;
    const authorizerLambda = new GoFunction(this, authorizerLambdaName, {
      functionName: authorizerLambdaName,
      entry: '../cmd/auth/lambda_authorizer/main.go',
      runtime: aws_lambda.Runtime.PROVIDED_AL2,
      timeout: config.lambda.Timeout,
      memorySize: config.lambda.MemorySize,
      logRetention: config.lambda.LogRetention,
      role: props.lambdaRole,
      bundling: config.lambda.GoBundling,
      environment: {
        JWT_SECRET_KEY,
        DDB_SESSIONS_TABLE_NAME: props.sessionsDB.tableName
      }
    });

    const authorizerName = config.AppName + '-Authorizer_' + props.stage;

    // Lambda Authorizer with 'REQUEST' type
    const authorizer = new apiGateway.RequestAuthorizer(this, authorizerName, {
      authorizerName,
      handler: authorizerLambda,
      identitySources: [apiGateway.IdentitySource.header('cookies')],
      resultsCacheTtl: Duration.minutes(5)
    });

    props.sessionsDB.grantReadWriteData(authorizerLambda);

    this.apiAuthorizer = authorizer;
  }
}
