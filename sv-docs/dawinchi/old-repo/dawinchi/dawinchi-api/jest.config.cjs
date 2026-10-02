module.exports = {
    preset: 'ts-jest',
    testEnvironment: 'node',
    transform: {},
    globals: {
      'ts-jest': {
        useESM: true
      }
    },
    extensionsToTreatAsEsm: ['.ts'],
    moduleNameMapper: {
      '^(\\.{1,2}/.*)\\.js$': '$1',
    },
    resolver: 'jest-ts-webcompat-resolver',
  };
  