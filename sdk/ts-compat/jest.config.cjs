/** @type {import('ts-jest').JestConfigWithTsJest} */
module.exports = {
    preset: 'ts-jest',
    testEnvironment: 'node',
    roots: ['<rootDir>/test/unit'],
    testMatch: ['**/test/unit/**/*.test.ts'],
    collectCoverageFrom: ['src/**/*.ts', '!src/**/index.ts', '!src/**/*.d.ts'],
    collectCoverage: true,
    coverageReporters: ['text'],
    transform: {
        '^.+\\.[jt]s$': ['ts-jest', { diagnostics: { ignoreDiagnostics: ['^'] } }],
    },
    // Allow ts-jest to transform ESM files from @yellow-org/sdk (file: dependency)
    transformIgnorePatterns: [
        'node_modules/(?!@yellow-org)',
    ],
    testTimeout: 10000,
    maxWorkers: '50%',
};
