import { readFileSync, existsSync } from 'fs';
import { join } from 'path';
import { exec } from 'child_process';
import { promisify } from 'util';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = join(__filename, '..');

const execAsync = promisify(exec);

interface ValidationResult {
    success: boolean;
    message: string;
    details?: string;
}

/**
 * Validates TypeScript compilation
 */
async function validateTypeScriptCompilation(): Promise<ValidationResult> {
    try {
        const { stdout, stderr } = await execAsync('npx tsc --noEmit', {
            cwd: join(__dirname, '..'),
        });

        if (stderr && stderr.includes('error')) {
            return {
                success: false,
                message: 'TypeScript compilation errors found',
                details: stderr,
            };
        }

        return {
            success: true,
            message: 'TypeScript compilation successful',
        };
    } catch (error) {
        return {
            success: false,
            message: 'TypeScript compilation failed',
            details: error instanceof Error ? error.message : 'Unknown compilation error',
        };
    }
}

/**
 * Validates that contract ABI files exist in the blockchain module
 */
async function validateContractSync(): Promise<ValidationResult> {
    try {
        // ABIs are now inlined in the blockchain/evm directory
        const abiFiles = [
            join(__dirname, '../src/blockchain/evm/channel_hub_abi.ts'),
            join(__dirname, '../src/blockchain/evm/erc20_abi.ts'),
        ];

        const missingFiles = abiFiles.filter((f) => !existsSync(f));

        if (missingFiles.length > 0) {
            return {
                success: false,
                message: 'Contract ABI files not found',
                details: `Missing: ${missingFiles.join(', ')}`,
            };
        }

        return {
            success: true,
            message: 'Contract ABI files are present',
        };
    } catch (error) {
        return {
            success: false,
            message: 'Error checking contract ABI files',
            details: error instanceof Error ? error.message : 'Unknown sync error',
        };
    }
}

/**
 * Validates SDK exports and structure
 */
async function validateSDKStructure(): Promise<ValidationResult> {
    try {
        const indexPath = join(__dirname, '../src/index.ts');
        const content = readFileSync(indexPath, 'utf-8');

        // Check for essential exports
        const essentialExports = [
            "'./client'",
            "'./signers'",
            "'./utils'",
            "'./core'",
            "'./blockchain'",
            "'./rpc'",
        ];

        const missingExports = essentialExports.filter((exp) => !content.includes(exp));

        if (missingExports.length > 0) {
            return {
                success: false,
                message: 'Missing essential SDK exports',
                details: `Missing: ${missingExports.join(', ')}`,
            };
        }

        return {
            success: true,
            message: 'SDK structure is valid',
        };
    } catch (error) {
        return {
            success: false,
            message: 'Error validating SDK structure',
            details: error instanceof Error ? error.message : 'Unknown structure error',
        };
    }
}

/**
 * Validates package.json configuration
 */
async function validatePackageConfig(): Promise<ValidationResult> {
    try {
        const packagePath = join(__dirname, '../package.json');
        const packageContent = JSON.parse(readFileSync(packagePath, 'utf-8'));

        // Check essential scripts
        const requiredScripts = ['build', 'typecheck', 'test', 'codegen-abi'];
        const missingScripts = requiredScripts.filter((script) => !packageContent.scripts[script]);

        if (missingScripts.length > 0) {
            return {
                success: false,
                message: 'Missing required npm scripts',
                details: `Missing scripts: ${missingScripts.join(', ')}`,
            };
        }

        // Check essential dependencies
        const requiredDeps = ['viem', 'abitype'];
        const missingDeps = requiredDeps.filter(
            (dep) => !packageContent.dependencies[dep] && !packageContent.devDependencies[dep],
        );

        if (missingDeps.length > 0) {
            return {
                success: false,
                message: 'Missing required dependencies',
                details: `Missing: ${missingDeps.join(', ')}`,
            };
        }

        return {
            success: true,
            message: 'Package configuration is valid',
        };
    } catch (error) {
        return {
            success: false,
            message: 'Error validating package configuration',
            details: error instanceof Error ? error.message : 'Unknown config error',
        };
    }
}

/**
 * Main validation function
 */
async function main() {
    console.log('Running SDK validation checks...\n');

    // Regenerate ABIs from latest contract artifacts before validating
    await execAsync('npm run codegen-abi', { cwd: join(__dirname, '..') });

    const validations = [
        { name: 'TypeScript Compilation', fn: validateTypeScriptCompilation },
        { name: 'Contract Sync', fn: validateContractSync },
        { name: 'SDK Structure', fn: validateSDKStructure },
        { name: 'Package Configuration', fn: validatePackageConfig },
    ];

    let allPassed = true;
    const results: { name: string; result: ValidationResult }[] = [];

    for (const validation of validations) {
        console.log(`Validating ${validation.name}...`);
        const result = await validation.fn();
        results.push({ name: validation.name, result });

        if (result.success) {
            console.log(`✅ ${validation.name}: ${result.message}`);
        } else {
            console.log(`❌ ${validation.name}: ${result.message}`);
            if (result.details) {
                console.log(`   Details: ${result.details}`);
            }
            allPassed = false;
        }
        console.log('');
    }

    // Summary
    console.log('📊 Validation Summary:');
    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

    results.forEach(({ name, result }) => {
        const status = result.success ? '✅' : '❌';
        console.log(`${status} ${name}: ${result.message}`);
    });

    console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

    if (allPassed) {
        console.log('🎉 All validation checks passed! SDK is reliable and ready.');
        process.exit(0);
    } else {
        console.log('💥 Some validation checks failed. Please fix the issues above.');
        process.exit(1);
    }
}

main();
