#!/usr/bin/env node
import 'source-map-support/register';
import { App } from 'aws-cdk-lib';
import { StatefulStack } from '../lib/stacks/stateful';

const app = new App();

// TODO - figure the multi region and environments deployment

new StatefulStack(app, 'StatefulStack', {
  terminationProtection: false,
  env: {
    account: process.env.AWS_ACCOUNT ?? '054037097197',
    region: process.env.AWS_REGION ?? 'ap-south-1'
  },
  stage: 'dev',
  appName: 'TabsFlow',
  googleClientId: 'foo',
  googleClientSecret: 'bar'
});

// synthesize stacks;
app.synth();
