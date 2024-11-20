#!/usr/bin/env node
import 'source-map-support/register';
import { App, Aspects, RemovalPolicy } from 'aws-cdk-lib';
import { ServiceStack } from '../lib/stacks/services/service-stack';
import { StatefulStack } from '../lib/stacks/stateful';

import { AwsSolutionsChecks } from 'cdk-nag';

import { config } from '../config';

const app = new App();

new StatefulStack(app, 'StatefulStack', {
  terminationProtection: false,
  env: {
    region: process.env.AWS_REGION,
    account: process.env.AWS_ACCOUNT_ID
  },
  stage: config.Env.DEPLOY_STAGE,
  removalPolicy: config.Env.DEPLOY_STAGE === config.Stage.Prod ? RemovalPolicy.RETAIN : RemovalPolicy.DESTROY
});

new ServiceStack(app, 'ServiceStack', {
  terminationProtection: false,
  env: {
    region: process.env.AWS_REGION,
    account: process.env.AWS_ACCOUNT_ID
  },
  stage: config.Env.DEPLOY_STAGE,
  removalPolicy: config.Env.DEPLOY_STAGE === config.Stage.Prod ? RemovalPolicy.RETAIN : RemovalPolicy.DESTROY
});

// synthesize stacks;
app.synth();
