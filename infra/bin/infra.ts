import 'source-map-support/register';

import { App, RemovalPolicy } from 'aws-cdk-lib';

import { config } from '../config';
import { ACMStack } from './../lib/stacks/acm/acm';
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
  stage,
  stackName: 'GithubOIDCStack',
  env: {
    region: process.env.AWS_REGION,
    account: process.env.AWS_ACCOUNT_ID
  },
  terminationProtection: stage === config.Stage.Prod
});

// acm stack
new ACMStack(app, 'ACMStack', {
  stage,
  stackName: 'ACMStack',
  env: {
    // API GW requires certificate to be in us-east-1 for custom domain (with edge endpoints)
    region: 'us-east-1',
    account: process.env.AWS_ACCOUNT_ID
  },
  terminationProtection: stage === config.Stage.Prod
});
