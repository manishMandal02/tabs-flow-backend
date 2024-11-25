import 'source-map-support/register';

import { App, RemovalPolicy } from 'aws-cdk-lib';

import { config } from '../config';

import { StatefulStack } from '../lib/stacks/stateful';
import { ServiceStack } from '../lib/stacks/services/services';
import { GithubOIDCStack } from '../lib/stacks/github-oidc/github-oidc';

const app = new App();

const stage = config.Env.DEPLOY_STAGE;

new StatefulStack(app, 'StatefulStack', {
  stage,
  stackName: 'StatefulStack',
  env: {
    region: process.env.AWS_REGION,
    account: process.env.AWS_ACCOUNT_ID
  },
  terminationProtection: stage === config.Stage.Prod,
  removalPolicy: stage === config.Stage.Prod ? RemovalPolicy.RETAIN : RemovalPolicy.DESTROY
});

new ServiceStack(app, 'ServiceStack', {
  stage,
  stackName: 'ServiceStack',
  env: {
    region: process.env.AWS_REGION,
    account: process.env.AWS_ACCOUNT_ID
  },
  terminationProtection: stage === config.Stage.Prod,
  removalPolicy: stage === config.Stage.Prod ? RemovalPolicy.RETAIN : RemovalPolicy.DESTROY
});

// github oidc stack
new GithubOIDCStack(app, 'GithubOIDCStack', {
  stackName: 'GithubOIDCStack',
  stage,
  env: {
    region: process.env.AWS_REGION,
    account: process.env.AWS_ACCOUNT_ID
  },
  terminationProtection: stage === config.Stage.Prod
});
