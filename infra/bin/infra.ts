#!/usr/bin/env node
import 'source-map-support/register';
import { App } from 'aws-cdk-lib';
import { ServiceStack } from '../lib/stacks/services/service-stack';
import { StatefulStack } from '../lib/stacks/stateful';

const app = new App();

// TODO: figure the multi region and environments deployment
const config = {
  stage: 'dev'
};

new StatefulStack(app, 'StatefulStack', {
  terminationProtection: false,
  env: {
    region: process.env.AWS_REGION
  },
  ...config
});
new ServiceStack(app, 'ServiceStack', {
  terminationProtection: false,
  ...config,
  env: {
    region: process.env.AWS_REGION
  }
});

// synthesize stacks;
app.synth();
