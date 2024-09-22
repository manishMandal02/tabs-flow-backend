#!/usr/bin/env node
import 'source-map-support/register';
import { App } from 'aws-cdk-lib';
import { ServiceStack } from '../lib/stacks/services/services-stack';
import { StatefulStack } from '../lib/stacks/stateful';

// TODO - Add TTL for Session Table while creating the table

const app = new App();

// TODO - figure the multi region and environments deployment
const config = {
  stage: 'dev'
};

const statefulStack = new StatefulStack(app, 'StatefulStack', {
  terminationProtection: false,
  env: {
    account: process.env.AWS_ACCOUNT,
    region: process.env.AWS_REGION
  },
  ...config
});

new ServiceStack(app, 'ServiceStack', {
  terminationProtection: false,
  ...config,
  env: {
    account: process.env.AWS_ACCOUNT,

    region: process.env.AWS_REGION
  },
  mainDB: statefulStack.mainDB,
  sessionsDB: statefulStack.sessionsDB
});

// synthesize stacks;
app.synth();
