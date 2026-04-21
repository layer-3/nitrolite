/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
    preset: 'ts-jest/presets/default-esm',
    testEnvironment: 'node',
    roots: ['<rootDir>/test/unit'],
    testMatch: ['**/test/unit/**/*.test.ts'],
    collectCoverageFrom: ['src/**/*.ts', '!src/**/index.ts', '!src/**/*.d.ts'],
    collectCoverage: true,
    coverageReporters: ['text'],
    extensionsToTreatAsEsm: ['.ts'],
    transform: {
        '^.+\\.ts$': [
            'ts-jest',
            {
                useESM: true,
                tsconfig: {
                    module: 'Node16',
                    moduleResolution: 'Node16',
                    types: ['node', 'jest'],
                },
            },
        ],
    },
    testTimeout: 10000, // 10 seconds for unit tests (should be fast)
    maxWorkers: '50%', // Can run unit tests in parallel
    moduleNameMapper: {
        '^@/(.*)$': '<rootDir>/src/$1',
        '^(\\.{1,2}/.*)\\.js$': '$1',
    },
};
