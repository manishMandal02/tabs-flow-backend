import { Construct } from 'constructs';

import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { aws_apigateway, aws_dynamodb, aws_iam } from 'aws-cdk-lib';

import { config } from '../../../config';

type NotesServiceProps = {
  stage: string;
  apiGW: aws_apigateway.RestApi;
  mainDB: aws_dynamodb.ITable;
  searchIndexDB: aws_dynamodb.ITable;
  lambdaRole: aws_iam.Role;
  apiAuthorizer: aws_apigateway.RequestAuthorizer;
};

export class NotesService extends Construct {
  constructor(scope: Construct, props: NotesServiceProps, id = 'NotesService') {
    super(scope, id);

    const notesServiceLambdaName = `${id}_${props.stage}`;
    const notesServiceLambda = new GoFunction(this, notesServiceLambdaName, {
      functionName: notesServiceLambdaName,
      entry: '../cmd/notes/main.go',
      runtime: config.lambda.Runtime,
      timeout: config.lambda.Timeout,
      memorySize: config.lambda.MemorySize,
      logRetention: config.lambda.LogRetention,
      role: props.lambdaRole,
      architecture: config.lambda.Architecture,
      bundling: config.lambda.GoBundling,
      environment: {
        DDB_MAIN_TABLE_NAME: props.mainDB.tableName,
        SEARCH_INDEX_TABLE_NAME: props.searchIndexDB.tableName
      }
    });

    // grant permissions to lambda to read/write to dynamodb and send message to email queue
    props.mainDB.grantReadWriteData(notesServiceLambda);
    props.searchIndexDB.grantReadWriteData(notesServiceLambda);

    // add notes resource/endpoints to api gateway
    const notesResource = props.apiGW.root.addResource('notes').addProxy({ anyMethod: false });
    notesResource.addMethod('ANY', new aws_apigateway.LambdaIntegration(notesServiceLambda), {
      authorizationType: aws_apigateway.AuthorizationType.CUSTOM,
      authorizer: props.apiAuthorizer
    });
  }
}
