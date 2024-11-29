import { Duration } from 'aws-cdk-lib';
import { Architecture, Runtime } from 'aws-cdk-lib/aws-lambda';
import { RetentionDays } from 'aws-cdk-lib/aws-logs';
import * as dotenv from 'dotenv';

dotenv.config({ path: '../.env' });

// helper to get env variables
const getEnv = (key: string) => {
  // api domain name is not required in test environment
  if (key === 'API_DOMAIN_NAME' && getEnv('DEPLOY_STAGE') === Stage.Test) {
    return '';
  }

  const evn = process.env[key];

  if (!evn) {
    throw new Error(`Missing env variable: ${key}`);
  }
  return evn;
};

const AppName = 'TabsFlow';

const Stage = {
  Dev: 'dev',
  Prod: 'prod',
  Test: 'test'
} as const;

const GithubOIDC = {
  domain: 'token.actions.githubusercontent.com',
  owner: 'manishMandal02',
  repo: 'tabs-flow-backend',
  roleName: 'GithubActionsDeployRole'
} as const;

const DynamoDB = {
  MainTableName: `${AppName}-Main_${getEnv('DEPLOY_STAGE')}`,
  SessionsTableName: `${AppName}-Sessions_${getEnv('DEPLOY_STAGE')}`,
  SearchIndexTableName: `${AppName}-SearchIndex_${getEnv('DEPLOY_STAGE')}`,
  PrimaryKey: 'PK',
  SortKey: 'SK',
  TTL: 'TTL'
} as const;

const ssmParamNameBase = `/${AppName.toLowerCase()}/${getEnv('DEPLOY_STAGE')}`;

const SSMParameterName = {
  MainTableArn: `${ssmParamNameBase}/main-table-arn`,
  SessionsTableArn: `${ssmParamNameBase}/sessions-table-arn`,
  SearchIndexTableArn: `${ssmParamNameBase}/search-index-table-arn`,
  APIDomainCertArn: `${ssmParamNameBase}/api-domain-cert-arn`
} as const;

const Lambda = {
  MemorySize: 128,
  Timeout: Duration.seconds(20),
  LogRetention: getEnv('DEPLOY_STAGE') === Stage.Prod ? RetentionDays.TWO_WEEKS : RetentionDays.ONE_DAY,
  Architecture: Architecture.ARM_64,
  Runtime: Runtime.PROVIDED_AL2,
  GoBundling: {
    // goBuildFlags: ['-ldflags="-s -w"']
  }
} as const;

const Env = {
  AWS_REGION: getEnv('AWS_REGION'),
  DEPLOY_STAGE: getEnv('DEPLOY_STAGE'),
  JWT_SECRET_KEY: getEnv('JWT_SECRET_KEY'),
  API_DOMAIN_NAME: getEnv('API_DOMAIN_NAME'),
  VAPID_PUBLIC_KEY: getEnv('VAPID_PUBLIC_KEY'),
  VAPID_PRIVATE_KEY: getEnv('VAPID_PRIVATE_KEY'),
  ZEPTO_MAIL_API_KEY: getEnv('ZEPTO_MAIL_API_KEY')
} as const;

const AllowedOrigins = [
  'https://localhost:3000',
  'https://localhost:3003',
  'https://tabsflow.com',
  'https://www.tabsflow.com'
];

export const config = {
  Env,
  Stage,
  AppName,
  Lambda,
  DynamoDB,
  GithubOIDC,
  AllowedOrigins,
  SSMParameterName
} as const;
