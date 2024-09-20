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

const common = {
  AppName: 'TabsFlow',
  DDB: dynamoDB
};

export const config = {
  Dev,
  Prod,
  ...common
};
