#!/usr/bin/env node
import 'source-map-support/register';
import { App } from 'aws-cdk-lib';
import { StatefulStack } from '../lib/stacks/stateful';
import { APIServiceStack } from '../lib/stacks/api-service/api-service';

const app = new App();

// TODO - figure the multi region and environments deployment
const config = {
  stage: 'dev',
  appName: 'TabsFlow',
  googleClientId: 'foo',
  googleClientSecret: 'bar',
  sesEmail: 'support@tabflows.com',
  emailQueueName: 'EmailQueue'
};

new StatefulStack(app, 'StatefulStack', {
  terminationProtection: false,
  env: {
    account: process.env.AWS_ACCOUNT ?? '054037097197',
    region: process.env.AWS_REGION ?? 'ap-south-1'
  },
  ...config
});

new APIServiceStack(app, 'APIServiceStack', {
  terminationProtection: false,
  env: {
    account: process.env.AWS_ACCOUNT ?? '054037097197',
    region: process.env.AWS_REGION ?? 'ap-south-1'
  },
  ...config
});

// synthesize stacks;
app.synth();
