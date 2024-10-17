import { Duration } from 'aws-cdk-lib';
import { Architecture, Runtime } from 'aws-cdk-lib/aws-lambda';
import { RetentionDays } from 'aws-cdk-lib/aws-logs';
import * as dotenv from 'dotenv';
import * as path from 'path';

dotenv.config({ path: path.resolve(__dirname, '../.env') });

const Stage = {
  Dev: 'dev',
  Prod: 'prod'
};

const dynamoDB = {
  MainTableName: 'Main',
  SessionsTableName: 'Sessions',
  SearchIndexTableName: 'SearchIndex',
  PrimaryKey: 'PK',
  SortKey: 'SK',
  TTL: 'TTL'
};

const lambda = {
  MemorySize: 128,
  Timeout: Duration.seconds(20),
  LogRetention: RetentionDays.TWO_WEEKS,
  Architecture: Architecture.ARM_64,
  Runtime: Runtime.PROVIDED_AL2,
  GoBundling: {
    // goBuildFlags: ['-ldflags="-s -w"']
  }
};

const common = {
  AppName: 'TabsFlow'
};

// helper to get env variables
const getEnv = (key: string) => {
  const evn = process.env[key];

  if (!evn) {
    throw new Error(`Missing env variable: ${key}`);
  }
  return evn;
};

const Env = {
  DEPLOY_STAGE: getEnv('DEPLOY_STAGE'),
  API_DOMAIN_NAME: getEnv('API_DOMAIN_NAME'),
  JWT_SECRET_KEY: getEnv('JWT_SECRET_KEY'),
  ZEPTO_MAIL_API_KEY: getEnv('ZEPTO_MAIL_API_KEY'),
  VAPID_PRIVATE_KEY: getEnv('VAPID_PRIVATE_KEY'),
  VAPID_PUBLIC_KEY: getEnv('VAPID_PUBLIC_KEY')
};

export const config = {
  Stage,
  Env,
  lambda,
  dynamoDB,
  ...common
};
