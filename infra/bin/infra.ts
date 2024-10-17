#!/usr/bin/env node
import 'source-map-support/register';
import { App } from 'aws-cdk-lib';
import { ServiceStack } from '../lib/stacks/services/service-stack';
import { StatefulStack } from '../lib/stacks/stateful';
import { config } from '../config';

const app = new App();

// TODO: figure the multi region and environments deployment

new StatefulStack(app, 'StatefulStack', {
  terminationProtection: false,
  env: {
    region: process.env.AWS_REGION,
    account: process.env.AWS_ACCOUNT_ID
  },
  stage: config.Env.DEPLOY_STAGE
});
new ServiceStack(app, 'ServiceStack', {
  terminationProtection: false,
  env: {
    region: process.env.AWS_REGION,
    account: process.env.AWS_ACCOUNT_ID
  },
  stage: config.Env.DEPLOY_STAGE
});

// synthesize stacks;
app.synth();
