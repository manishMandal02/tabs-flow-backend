/** @type {import('ts-jest').JestConfigWithTsJest} **/
export default {
  testEnvironment: 'node',
  transform: {
    '^.+.tsx?$': ['ts-jest', {}]
  },
  testEnvironmentOptions: {
    NODE_OPTIONS: '--experimental-vm-modules'
  }
};
