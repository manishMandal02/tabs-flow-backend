import { Construct } from 'constructs';

import { aws_apigateway, aws_dynamodb, aws_iam, Duration, aws_sqs as sqs } from 'aws-cdk-lib';
import { config } from '../../../config';
import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import * as eventSources from 'aws-cdk-lib/aws-lambda-event-sources';

type NotificationsServicePops = {
  stage: string;
  db: aws_dynamodb.ITable;
  lambdaRole: aws_iam.Role;
  emailQueue: sqs.Queue;
  apiGW: aws_apigateway.RestApi;
  apiAuthorizer: aws_apigateway.RequestAuthorizer;
};

export class NotificationsService extends Construct {
  Queue: sqs.Queue;
  constructor(scope: Construct, props: NotificationsServicePops, id = 'NotificationService') {
    super(scope, id);

    const queueName = `${config.AppName}-Notifications_${props.stage}`;

    // sqs queue
    const dlqNotifications = new sqs.Queue(this, queueName + '-dlq', {
      visibilityTimeout: Duration.seconds(300)
    });

    const notificationsQueue = new sqs.Queue(this, queueName, {
      queueName,
      visibilityTimeout: Duration.seconds(300),
      deliveryDelay: Duration.seconds(1),
      deadLetterQueue: {
        queue: dlqNotifications,
        maxReceiveCount: 3
      }
    });

    const schedulerRole = new aws_iam.Role(this, 'SchedulerRole', {
      assumedBy: new aws_iam.ServicePrincipal('scheduler.amazonaws.com')
    });
    schedulerRole.addToPolicy(
      new aws_iam.PolicyStatement({
        actions: ['sqs:SendMessage'],
        resources: [notificationsQueue.queueArn]
      })
    );

    const notificationsServiceLambdaName = `${id}_${props.stage}`;

    const notificationsServiceLambda = new GoFunction(this, notificationsServiceLambdaName, {
      functionName: notificationsServiceLambdaName,
      entry: '../cmd/notifications/main.go',
      runtime: config.lambda.Runtime,
      timeout: config.lambda.Timeout,
      memorySize: config.lambda.MemorySize,
      logRetention: config.lambda.LogRetention,
      role: props.lambdaRole,
      architecture: config.lambda.Architecture,
      bundling: config.lambda.GoBundling,
      environment: {
        DDB_MAIN_TABLE_NAME: props.db.tableName,
        NOTIFICATIONS_QUEUE_ARN: notificationsQueue.queueArn,
        SCHEDULER_ROLE_ARN: schedulerRole.roleArn,
        EMAIL_QUEUE_URL: props.emailQueue.queueUrl,
        VAPID_PRIVATE_KEY: config.Env.VAPID_PRIVATE_KEY,
        VAPID_PUBLIC_KEY: config.Env.VAPID_PUBLIC_KEY
      }
    });

    notificationsQueue.grantConsumeMessages(notificationsServiceLambda);

    notificationsServiceLambda.addEventSource(
      new eventSources.SqsEventSource(notificationsQueue, {
        batchSize: 1
      })
    );

    props.emailQueue.grantSendMessages(notificationsServiceLambda);

    // grant permissions to lambda to read/write to dynamodb and send message to email queue
    props.db.grantReadWriteData(notificationsServiceLambda);

    // add users resource/endpoints to api gateway
    const usersResource = props.apiGW.root.addResource('notifications').addProxy({ anyMethod: false });
    usersResource.addMethod('ANY', new aws_apigateway.LambdaIntegration(notificationsServiceLambda), {
      authorizationType: aws_apigateway.AuthorizationType.CUSTOM,
      authorizer: props.apiAuthorizer
    });

    this.Queue = notificationsQueue;
  }
}
