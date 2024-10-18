import { Construct } from 'constructs';

import { aws_apigateway, aws_dynamodb, aws_iam as iam, Duration, aws_sqs as sqs, Stack } from 'aws-cdk-lib';
import { config } from '../../../config';
import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import * as eventSources from 'aws-cdk-lib/aws-lambda-event-sources';

type NotificationsServicePops = {
  stage: string;
  db: aws_dynamodb.ITable;
  lambdaRole: iam.Role;
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

    // Create the scheduler execution role
    const schedulerExecutionRole = new iam.Role(this, 'NotificationServiceSchedulerRole', {
      assumedBy: new iam.ServicePrincipal('scheduler.amazonaws.com'),
      description: 'Role that EventBridge Scheduler can assume to execute tasks'
    });

    // Add permissions to the scheduler execution role
    schedulerExecutionRole.addToPolicy(
      new iam.PolicyStatement({
        effect: iam.Effect.ALLOW,
        actions: [
          'lambda:InvokeFunction'
          // Add other permissions needed by your scheduled tasks
        ],
        resources: [
          // Specify the exact Lambda ARN that the scheduler needs to invoke
          `arn:aws:lambda:${Stack.of(this).region}:${Stack.of(this).account}:function:*`
        ]
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
        SCHEDULER_ROLE_ARN: schedulerExecutionRole.roleArn,
        EMAIL_QUEUE_URL: props.emailQueue.queueUrl,
        VAPID_PRIVATE_KEY: config.Env.VAPID_PRIVATE_KEY,
        VAPID_PUBLIC_KEY: config.Env.VAPID_PUBLIC_KEY
      }
    });

    // create a policy for EventBridge Scheduler permissions
    const schedulerPolicy = new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: [
        'scheduler:CreateSchedule',
        'scheduler:DeleteSchedule',
        'scheduler:GetSchedule',
        'scheduler:UpdateSchedule',
        'scheduler:ListSchedules',
        // required  to assume roles
        'iam:PassRole'
      ],
      resources: [
        // Replace with your region and account
        `arn:aws:scheduler:${Stack.of(this).region}:${Stack.of(this).account}:schedule/*`
      ]
    });

    // Create a specific PassRole policy
    const passRolePolicy = new iam.PolicyStatement({
      effect: iam.Effect.ALLOW,
      actions: ['iam:PassRole'],
      resources: [
        // Specifically allow passing the scheduler execution role
        schedulerExecutionRole.roleArn
      ]
    });

    // lambda permissions
    notificationsServiceLambda.addToRolePolicy(passRolePolicy);

    notificationsServiceLambda.addToRolePolicy(schedulerPolicy);

    notificationsServiceLambda.addEventSource(
      new eventSources.SqsEventSource(notificationsQueue, {
        batchSize: 1
      })
    );

    // sqs queue permissions
    notificationsQueue.grantConsumeMessages(notificationsServiceLambda);
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
