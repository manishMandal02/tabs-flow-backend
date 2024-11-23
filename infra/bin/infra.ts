#!/usr/bin/env node
import 'source-map-support/register';
import { App, RemovalPolicy } from 'aws-cdk-lib';
import { ServiceStack } from '../lib/stacks/services/service-stack';
import { StatefulStack } from '../lib/stacks/stateful';

import { config } from '../config';

const app = new App();

const stage = config.Env.DEPLOY_STAGE;

new StatefulStack(app, 'StatefulStack', {
  terminationProtection: stage === config.Stage.Prod,
  env: {
    region: process.env.AWS_REGION,
    account: process.env.AWS_ACCOUNT_ID
  },
  stage: stage,
  removalPolicy: stage === config.Stage.Prod ? RemovalPolicy.RETAIN : RemovalPolicy.DESTROY
});

new ServiceStack(app, 'ServiceStack', {
  terminationProtection: stage === config.Stage.Prod,
  env: {
    region: process.env.AWS_REGION,
    account: process.env.AWS_ACCOUNT_ID
  },
  stage: stage,
  removalPolicy: stage === config.Stage.Prod ? RemovalPolicy.RETAIN : RemovalPolicy.DESTROY
});
