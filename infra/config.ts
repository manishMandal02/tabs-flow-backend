import { Duration } from 'aws-cdk-lib';
import { Architecture, Runtime } from 'aws-cdk-lib/aws-lambda';
import { RetentionDays } from 'aws-cdk-lib/aws-logs';
import * as dotenv from 'dotenv';
import * as path from 'path';

dotenv.config({ path: path.resolve(__dirname, '../.env') });

const AppName = 'TabsFlow';

const Stage = {
  Dev: 'dev',
  Prod: 'prod',
  Test: 'test'
} as const;

// helper to get env variables
const getEnv = (key: string) => {
  const evn = process.env[key];

  if (!evn) {
    throw new Error(`Missing env variable: ${key}`);
  }
  return evn;
};

const dynamoDB = {
  MainTableName: `${AppName}-Main_${getEnv('DEPLOY_STAGE')}`,
  SessionsTableName: `${AppName}-Sessions_${getEnv('DEPLOY_STAGE')}`,
  SearchIndexTableName: `${AppName}-SearchIndex_${getEnv('DEPLOY_STAGE')}`,
  PrimaryKey: 'PK',
  SortKey: 'SK',
  TTL: 'TTL'
} as const;

const ssmParamNameBase = `/${AppName.toLowerCase()}/${getEnv('DEPLOY_STAGE')}`;

const ssmParameterNames = {
  MainTableArn: `${ssmParamNameBase}/main-table-arn`,
  SessionsTableArn: `${ssmParamNameBase}/sessions-table-arn`,
  SearchIndexTableArn: `${ssmParamNameBase}/search-index-table-arn`
} as const;

const lambda = {
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
  DEPLOY_STAGE: getEnv('DEPLOY_STAGE'),
  API_DOMAIN_NAME: getEnv('API_DOMAIN_NAME'),
  JWT_SECRET_KEY: getEnv('JWT_SECRET_KEY'),
  ZEPTO_MAIL_API_KEY: getEnv('ZEPTO_MAIL_API_KEY'),
  VAPID_PRIVATE_KEY: getEnv('VAPID_PRIVATE_KEY'),
  VAPID_PUBLIC_KEY: getEnv('VAPID_PUBLIC_KEY')
} as const;

export const config = {
  AppName,
  Stage,
  Env,
  Lambda: lambda,
  DynamoDB: dynamoDB,
  SSMParameterName: ssmParameterNames
} as const;
