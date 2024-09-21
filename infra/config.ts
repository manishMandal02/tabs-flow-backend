import { Duration } from 'aws-cdk-lib';

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
  Timeout: Duration.seconds(20),
  MemorySize: 128
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
  AWS_REGION: getEnv('AWS_REGION'),
  JWT_SECRET_KEY: getEnv('JWT_SECRET_KEY'),
  ZEPTO_MAIL_API_KEY: getEnv('ZEPTO_MAIL_API_KEY'),
  EMAIL_SQS_QUEUE_URL: getEnv('EMAIL_SQS_QUEUE_URL'),
  DDB_MAIN_TABLE_NAME: getEnv('DDB_MAIN_TABLE_NAME'),
  DDB_SESSIONS_TABLE_NAME: getEnv('DDB_SESSIONS_TABLE_NAME')
};

export const config = {
  Dev,
  Prod,
  Env,
  lambda,
  dynamoDB,
  ...common
};
