import { Construct } from 'constructs';

import * as iam from 'aws-cdk-lib/aws-iam';
import { Duration, Stack, StackProps } from 'aws-cdk-lib';

import { config } from '../../../config';

type GithubOIDCStackProps = StackProps & {
  stage: string;
};

export class GithubOIDCStack extends Stack {
  constructor(scope: Construct, id: string, props: GithubOIDCStackProps) {
    super(scope, id, props);

    // create an openid connect provider for github
    const ghProvider = new iam.OpenIdConnectProvider(this, 'githubProvider', {
      url: `https://${config.GithubOIDC.domain}`,
      clientIds: ['sts.amazonaws.com']
    });

    // grant only requests coming from a specific GitHub repository
    const conditions: iam.Conditions = {
      StringLike: {
        [`${config.GithubOIDC.domain}:sub`]: `repo:${config.GithubOIDC.owner}/${config.GithubOIDC.repo}:*`,
        [`${config.GithubOIDC.domain}:aud`]: 'sts.amazonaws.com'
      }
    };

    //  policy that allows actions on all necessary resources in the account
    const policy = new iam.PolicyDocument({
      statements: [
        new iam.PolicyStatement({
          effect: iam.Effect.ALLOW,
          actions: [
            'cloudformation:*',
            's3:*',
            'lambda:*',
            'apigateway:*',
            'dynamodb:*',
            'sqs:*',
            'scheduler:*',
            'events:*',
            'logs:*',
            'acm:*',
            'cloudwatch:*',
            'ssm:*',
            'ecr:*',
            'ec2:*',
            'iam:*'
          ],
          resources: ['*']
        })
      ]
    });

    // create a role that can be assumed by GitHub Actions
    new iam.Role(this, config.GithubOIDC.roleName, {
      assumedBy: new iam.WebIdentityPrincipal(ghProvider.openIdConnectProviderArn, conditions),
      roleName: config.GithubOIDC.roleName,
      description: 'This role is used via GitHub Actions to deploy with AWS CDK to the target AWS account',
      maxSessionDuration: Duration.hours(1),
      inlinePolicies: { 'github-actions-deployment-policy': policy }
    });
  }
}
