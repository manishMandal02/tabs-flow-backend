import { Duration } from 'aws-cdk-lib';
import { RetentionDays } from 'aws-cdk-lib/aws-logs';
import * as dotenv from 'dotenv';
import * as path from 'path';

dotenv.config({ path: path.resolve(__dirname, '../.env') });

const Dev = {
  Stage: 'dev'
};

const Prod = {
  Stage: 'prod'
};

const dynamoDB = {
  MainTableName: 'TabsFlow',
  SessionsTable: 'Sessions',
  PrimaryKey: 'PK',
  SortKey: 'SK',
  TTL: 'TTL'
};

const lambda = {
  MemorySize: 128,
  Timeout: Duration.seconds(20),
  LogRetention: RetentionDays.ONE_MONTH,
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
  JWT_SECRET_KEY: getEnv('JWT_SECRET_KEY'),
  ZEPTO_MAIL_API_KEY: getEnv('ZEPTO_MAIL_API_KEY')
};

export const config = {
  Dev,
  Prod,
  Env,
  lambda,
  dynamoDB,
  ...common
};
