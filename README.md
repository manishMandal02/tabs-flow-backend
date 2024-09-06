# TabsFlow Backend

> _AWS Serverless_ based architecture with **Golang** as core language for services and **Typescript** for (CDK) infra code.

## Technologies

- Golang: services
- TypeScript: CDK infra
- AWS: cloud services
  - CDK: IaC 
  - APIGateway: Rest APIs
  - Cognito:  authN & authZ
  - Lambda: compute services
  - DynamoDB: main data store
  - SQS: async tasks
  - SES: transactional emails
  - CloudWatch: monitoring & logging
  - EventBridge: scheduling future tasks

## Services

- **Auth**: authentication and authorization
  - Google OAuth
  - APIGateway Authorizer
  - Email based OTP verification
  - JWT session management
- **User**: user management
  - _Get/PUT_: profile
  - _Get/PUT_: preference
  - _Get/PUT_: subscription
- **Spaces**: tabs & spaces management
  - _GET/PUT_: spaces
  - _GET/PUT_: tabs
  - _GET/PUT_: groups
- **Sync**: sync local data to cloud
  - _POST_: sync data
- **UsageAnalytics**: analytics service
  - _GET_: get space usage data
  - _PATCH_: update user space usage
- **Notification**: notification management
  - _GET_: get notifications
  - _POST_: create notification
  - _DELETE_: delete notification
- **Email**: send transactional emails
  - !no direct api access
  - _Polls a SQS queue_ (Email Queue) \
    for message to send emails
- **Scheduler**: schedules tasks for future
  - !no direct api access
  - _Polls a SQS queue_ (Scheduler Queue) \
    for message to schedule tasks \
    (e.g. note remainder)
- **Monitoring**: monitoring & observability
  - services sends metrics to CloudWatch
  - set up alerts/alarms for metrics
  - structured logging (sent to CW)
