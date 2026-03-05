import { readFileSync, writeFileSync, readdirSync, statSync } from 'fs';
import { join, extname } from 'path';
import { glob } from 'glob';

interface ContractInfo {
    name: string;
    abi: any[];
    functions: FunctionInfo[];
    events: EventInfo[];
    errors: ErrorInfo[];
}

interface FunctionInfo {
    name: string;
    inputs: any[];
    outputs: any[];
    stateMutability: string;
    description?: string;
    businessContext?: string;
    usageExamples?: string[];
    relatedFunctions?: string[];
}

interface EventInfo {
    name: string;
    inputs: any[];
    description?: string;
}

interface ErrorInfo {
    name: string;
    inputs: any[];
    description?: string;
}

interface JSDocInfo {
    description: string;
    params: { [key: string]: string };
    returns: string;
    examples: string[];
}

/**
 * Extracts JSDoc comments for a specific function from TypeScript files
 */
function extractJSDocForFunction(functionName: string): JSDocInfo | null {
    const srcDir = join(__dirname, '../src');
    const tsFiles = glob.sync('**/*.ts', { cwd: srcDir });

    for (const file of tsFiles) {
        try {
            const content = readFileSync(join(srcDir, file), 'utf-8');

            // Look for JSDoc comments immediately before function declaration
            const functionRegex = new RegExp(
                `\\/\\*\\*([\\s\\S]*?)\\*\\/\\s*(?:async\\s+)?(?:public\\s+|private\\s+|protected\\s+)?${functionName}\\s*\\(`,
                'g',
            );

            const match = functionRegex.exec(content);
            if (match) {
                const jsdocContent = match[1];

                // Skip if this looks like a class-level comment (contains "class" or generic terms)
                if (
                    jsdocContent.toLowerCase().includes('class') ||
                    jsdocContent.toLowerCase().includes('main client') ||
                    jsdocContent.toLowerCase().includes('for interacting with')
                ) {
                    continue;
                }

                // Extract description (first meaningful line)
                const lines = jsdocContent.split('\n');
                let description = '';
                for (const line of lines) {
                    const cleaned = line.replace(/^\s*\*\s?/, '').trim();
                    if (cleaned && !cleaned.startsWith('@') && cleaned.length > 10) {
                        description = cleaned;
                        break;
                    }
                }

                // Extract @param annotations
                const params: { [key: string]: string } = {};
                const paramMatches = jsdocContent.matchAll(/\*\s*@param\s+(\w+)\s+(.+?)(?:\n|$)/g);
                for (const paramMatch of paramMatches) {
                    params[paramMatch[1]] = paramMatch[2].trim();
                }

                // Extract @returns annotation
                const returnsMatch = jsdocContent.match(/\*\s*@returns\s+(.+?)(?:\n|$)/);
                const returns = returnsMatch ? returnsMatch[1].trim() : '';

                if (description || Object.keys(params).length > 0 || returns) {
                    return { description, params, returns, examples: [] };
                }
            }
        } catch (error) {
            // Skip files that can't be read
        }
    }

    return null;
}

/**
 * Extracts business context from contract source files
 */
function extractContractContext(functionName: string): string | null {
    const contractDir = join(__dirname, '../../contract/src');

    if (!require('fs').existsSync(contractDir)) {
        return null;
    }

    try {
        const solFiles = glob.sync('**/*.sol', { cwd: contractDir });

        for (const file of solFiles) {
            try {
                const content = readFileSync(join(contractDir, file), 'utf-8');

                // Look for NatSpec comments immediately before function
                const natspecRegex = new RegExp(
                    `\\/\\*\\*([\\s\\S]*?)\\*\\/\\s*function\\s+${functionName}\\s*\\(`,
                    'g',
                );

                const natspecMatch = natspecRegex.exec(content);
                if (natspecMatch) {
                    const natspecContent = natspecMatch[1];
                    const lines = natspecContent.split('\n');

                    // Extract the main description (not @param, @return, etc.)
                    for (const line of lines) {
                        const cleaned = line.replace(/^\s*\*\s?/, '').trim();
                        if (
                            cleaned &&
                            !cleaned.startsWith('@') &&
                            !cleaned.startsWith('SPDX') &&
                            !cleaned.startsWith('pragma') &&
                            cleaned.length > 15 &&
                            !cleaned.toLowerCase().includes('license') &&
                            !cleaned.toLowerCase().includes('pragma')
                        ) {
                            return cleaned;
                        }
                    }
                }

                // Look for single-line comments immediately before function
                const singleCommentRegex = new RegExp(`\\/\\/\\s*(.+?)\\n\\s*function\\s+${functionName}\\s*\\(`, 'g');

                const singleMatch = singleCommentRegex.exec(content);
                if (singleMatch) {
                    const comment = singleMatch[1].trim();
                    if (
                        comment.length > 15 &&
                        !comment.toLowerCase().includes('todo') &&
                        !comment.toLowerCase().includes('fixme')
                    ) {
                        return comment;
                    }
                }
            } catch (error) {
                // Skip files that can't be read
            }
        }
    } catch (error) {
        // Skip if contract directory doesn't exist
    }

    return null;
}

/**
 * Infers function purpose from name and parameters
 */
function inferFunctionPurpose(functionName: string, inputs: any[], outputs: any[], stateMutability: string): string {
    const name = functionName.toLowerCase();

    // State channel specific functions
    if (name === 'challenge') {
        return 'Initiates a challenge against a state channel, disputing the current state with evidence of a more recent valid state.';
    }

    if (name === 'checkpoint') {
        return 'Updates the channel to a new agreed-upon state, typically used to progress the channel without disputes.';
    }

    if (name === 'close') {
        return 'Finalizes and closes a state channel, distributing the final balances to participants.';
    }

    if (name === 'create') {
        return 'Creates a new state channel with initial state and participant configuration.';
    }

    if (name === 'deposit') {
        return 'Deposits tokens or ETH into the custody contract for use in state channels.';
    }

    if (name === 'withdraw') {
        return "Withdraws available funds from the custody contract back to the user's wallet.";
    }

    if (name === 'join') {
        return 'Allows a participant to join an existing state channel by providing their signature.';
    }

    if (name.includes('account')) {
        if (outputs.length > 0) {
            return 'Retrieves account-specific information and balances for state channel operations.';
        }
        return 'Updates account-specific settings or state.';
    }

    if (name.includes('info') || name.includes('get')) {
        return 'Retrieves information about the contract or specific state channel data.';
    }

    if (name.includes('adjudicate')) {
        return 'Determines the validity of a state transition or resolves disputes between channel participants.';
    }

    if (name.includes('compare')) {
        return 'Compares two states to determine which is more recent or valid.';
    }

    // Generic patterns
    if (stateMutability === 'view' || stateMutability === 'pure') {
        if (outputs.length > 0) {
            return `Retrieves ${name} data without modifying contract state.`;
        }
        return `Validates or computes ${name} without modifying contract state.`;
    }

    if (stateMutability === 'payable') {
        return `Executes ${name} operation that may involve ETH transfers.`;
    }

    return `Executes ${name} operation on the contract state.`;
}

/**
 * Validates if extracted contract context is meaningful
 */
function isUsefulContractContext(context: string): boolean {
    if (!context || context.length < 20) return false;

    // Filter out common non-useful patterns
    const unhelpfulPatterns = [
        'interface ',
        'contract ',
        'pragma ',
        'spdx',
        'license',
        'import ',
        'using ',
        'library ',
        '{',
        '}',
        '/*',
        '*/',
        '//',
        'error ',
        'struct ',
        'enum ',
        'uint256',
        'address',
        'mapping',
    ];

    const lowerContext = context.toLowerCase();

    // If it contains too many unhelpful patterns, it's probably not a function comment
    const unhelpfulCount = unhelpfulPatterns.filter((pattern) => lowerContext.includes(pattern.toLowerCase())).length;

    if (unhelpfulCount > 2) return false;

    // If it looks like a meaningful comment
    const goodPatterns = [
        'allows',
        'creates',
        'updates',
        'retrieves',
        'executes',
        'finalizes',
        'initiates',
        'validates',
        'deposits',
        'withdraws',
        'returns',
        'checks',
        'verifies',
        'sets',
        'gets',
    ];

    const hasGoodPattern = goodPatterns.some((pattern) => lowerContext.includes(pattern));

    return hasGoodPattern || (context.length > 30 && !lowerContext.includes('solidity'));
}

/**
 * Finds usage examples of a function in the examples directory
 */
function findUsageExamples(functionName: string): string[] {
    const examples: string[] = [];
    const examplesDir = join(__dirname, '../../examples');

    if (!require('fs').existsSync(examplesDir)) {
        return examples;
    }

    try {
        const files = glob.sync('**/*.{ts,js,vue}', { cwd: examplesDir });

        for (const file of files) {
            try {
                const content = readFileSync(join(examplesDir, file), 'utf-8');

                // Look for function calls with context
                const functionCallRegex = new RegExp(
                    `([\\s\\S]{0,200})\\.${functionName}\\s*\\([\\s\\S]{0,200}\\)([\\s\\S]{0,100})`,
                    'g',
                );

                let match;
                while ((match = functionCallRegex.exec(content)) !== null) {
                    const before = match[1].trim();
                    const after = match[2].trim();

                    // Extract the actual function call line
                    const lines = content.split('\n');
                    const callLineIndex = content.substring(0, match.index).split('\n').length - 1;
                    const callLine = lines[callLineIndex];

                    if (callLine && callLine.includes(functionName)) {
                        // Try to extract meaningful context
                        const contextLines = lines.slice(Math.max(0, callLineIndex - 2), callLineIndex + 3);
                        const contextBlock = contextLines.join('\n').trim();

                        if (contextBlock.length > 10 && !examples.find((ex) => ex.includes(callLine.trim()))) {
                            examples.push(contextBlock);
                        }
                    }
                }
            } catch (error) {
                // Skip files that can't be read
            }
        }
    } catch (error) {
        // Skip if examples directory doesn't exist or can't be read
    }

    return examples.slice(0, 3); // Limit to 3 examples
}

/**
 * Analyzes parameter usage to provide better descriptions
 */
function analyzeParameterUsage(functionName: string, params: any[]): { [key: string]: string } {
    const paramDescriptions: { [key: string]: string } = {};

    // Common parameter patterns and their human-readable descriptions
    const commonPatterns: { [key: string]: string } = {
        token: 'The token contract address (use address(0) for ETH)',
        amount: "Amount in the token's smallest unit (wei for ETH, etc.)",
        channelId: 'Unique identifier for the state channel',
        candidate: 'The proposed new state for the channel',
        proofs: 'Supporting states that prove the transition is valid',
        initial: 'The starting state when creating a new channel',
        user: 'Ethereum address of the user account',
        account: 'Ethereum address to query information for',
        destination: 'Where tokens should be sent',
        participants: 'Array of Ethereum addresses participating in the channel',
    };

    params.forEach((param) => {
        const paramName = param.name.toLowerCase();

        // Check for exact matches first
        if (commonPatterns[paramName]) {
            paramDescriptions[param.name] = commonPatterns[paramName];
            return;
        }

        // Check for partial matches
        for (const [pattern, description] of Object.entries(commonPatterns)) {
            if (paramName.includes(pattern) || pattern.includes(paramName)) {
                paramDescriptions[param.name] = description;
                break;
            }
        }

        // Fallback to type-based description
        if (!paramDescriptions[param.name]) {
            if (param.type === 'address') {
                paramDescriptions[param.name] = 'Ethereum address';
            } else if (param.type === 'uint256') {
                paramDescriptions[param.name] = 'Numeric value (in smallest units)';
            } else if (param.type === 'bytes32') {
                paramDescriptions[param.name] = 'Unique 32-byte identifier';
            } else if (param.type === 'bytes') {
                paramDescriptions[param.name] = 'Encoded data payload';
            } else {
                paramDescriptions[param.name] = `${param.type} value`;
            }
        }
    });

    return paramDescriptions;
}

/**
 * Generates enhanced parameter documentation with business context
 */
function generateEnhancedParameterDocs(params: any[], jsdocInfo?: JSDocInfo | null, functionName?: string): string {
    if (params.length === 0) return 'None';

    const paramDescriptions = functionName ? analyzeParameterUsage(functionName, params) : {};

    return params
        .map((param) => {
            // Use JSDoc description if available
            const jsdocDesc = jsdocInfo?.params[param.name];
            if (jsdocDesc) {
                return `- **\`${param.name}\`**: ${jsdocDesc}`;
            }

            // Use analyzed description
            const analyzedDesc = paramDescriptions[param.name];
            if (analyzedDesc) {
                return `- **\`${param.name}\`**: ${analyzedDesc}`;
            }

            // Fallback to basic type info
            const typeStr =
                param.type === 'tuple' && param.components
                    ? `${param.type} (${param.components.map((c: any) => `${c.name}: ${c.type}`).join(', ')})`
                    : param.type;
            return `- **\`${param.name}\`** (\`${typeStr}\`): ${param.internalType || 'Parameter value'}`;
        })
        .join('\n');
}

/**
 * Generates enhanced function documentation with business context
 */
function generateEnhancedFunctionDocs(functions: FunctionInfo[]): string {
    if (functions.length === 0) return '## Functions\n\nNo functions available.\n';

    let docs = '## Functions\n\n';

    functions.forEach((func) => {
        // Extract JSDoc information (skip generic ones)
        const jsdocInfo = extractJSDocForFunction(func.name);

        // Extract contract-level context
        const contractContext = extractContractContext(func.name);

        // Find usage examples
        const usageExamples = findUsageExamples(func.name);

        docs += `### \`${func.name}\`\n\n`;

        // Use the best available description
        let description = '';
        if (jsdocInfo?.description && jsdocInfo.description.length > 20) {
            description = jsdocInfo.description;
        } else if (contractContext && isUsefulContractContext(contractContext)) {
            description = contractContext;
        } else {
            description = inferFunctionPurpose(func.name, func.inputs, func.outputs, func.stateMutability);
        }

        docs += `${description}\n\n`;

        // Add technical details
        docs += `**Type:** \`${func.stateMutability}\`\n\n`;

        // Enhanced parameters
        docs += '**Parameters:**\n';
        docs += generateEnhancedParameterDocs(func.inputs, jsdocInfo, func.name) + '\n\n';

        // Enhanced returns
        if (func.outputs.length > 0) {
            docs += '**Returns:**\n';
            if (jsdocInfo?.returns) {
                docs += `${jsdocInfo.returns}\n\n`;
            } else {
                docs += generateEnhancedParameterDocs(func.outputs) + '\n\n';
            }
        }

        // Add usage examples from codebase
        if (usageExamples.length > 0) {
            docs += '**Real Usage Examples:**\n\n';
            usageExamples.forEach((example, index) => {
                docs += '```typescript\n';
                docs += example;
                docs += '\n```\n\n';
            });
        } else {
            // Generate a contextual example
            docs += '**Example Usage:**\n\n';
            docs += '```typescript\n';

            // Generate more meaningful examples based on function type
            if (func.stateMutability === 'view' || func.stateMutability === 'pure') {
                docs += `// Read ${func.name} from contract\n`;
                docs += `const result = await publicClient.readContract({\n`;
            } else {
                docs += `// Execute ${func.name} transaction\n`;
                docs += `const { request } = await publicClient.simulateContract({\n`;
            }

            docs += `  address: contractAddress,\n`;
            docs += `  abi: contractAbi,\n`;
            docs += `  functionName: '${func.name}',\n`;
            if (func.inputs.length > 0) {
                docs += `  args: [${func.inputs.map((i) => `${i.name}`).join(', ')}],\n`;
            }
            docs += `});\n`;

            if (func.stateMutability !== 'view' && func.stateMutability !== 'pure') {
                docs += `\n// Execute the transaction\n`;
                docs += `const hash = await walletClient.writeContract(request);\n`;
            }

            docs += '```\n\n';
        }
    });

    return docs;
}

/**
 * Extracts contract information from generated.ts
 */
function extractContractInfo(): ContractInfo[] {
    const generatedPath = join(__dirname, '../src/generated.ts');
    const content = readFileSync(generatedPath, 'utf-8');

    const contracts: ContractInfo[] = [];
    const abiExportRegex = /export const (\w+)Abi = (\[[\s\S]*?\]) as const/g;

    let match;
    while ((match = abiExportRegex.exec(content)) !== null) {
        const [, contractName, abiContent] = match;

        try {
            // Convert TypeScript format to valid JSON
            const jsonContent = abiContent
                .replace(/'/g, '"') // Replace single quotes with double quotes
                .replace(/(\w+):/g, '"$1":') // Quote property names
                .replace(/,(\s*[}\]])/g, '$1'); // Remove trailing commas

            const abi = JSON.parse(jsonContent);
            const functions: FunctionInfo[] = [];
            const events: EventInfo[] = [];
            const errors: ErrorInfo[] = [];

            abi.forEach((item: any) => {
                switch (item.type) {
                    case 'function':
                        functions.push({
                            name: item.name,
                            inputs: item.inputs || [],
                            outputs: item.outputs || [],
                            stateMutability: item.stateMutability,
                        });
                        break;
                    case 'event':
                        events.push({
                            name: item.name,
                            inputs: item.inputs || [],
                        });
                        break;
                    case 'error':
                        errors.push({
                            name: item.name,
                            inputs: item.inputs || [],
                        });
                        break;
                }
            });

            contracts.push({
                name: contractName.replace('Abi', ''),
                abi,
                functions,
                events,
                errors,
            });
        } catch (error) {
            console.warn(`Failed to parse ABI for ${contractName}:`, error);
        }
    }

    return contracts;
}

/**
 * Generates event documentation
 */
function generateEventDocs(events: EventInfo[]): string {
    if (events.length === 0) return '## Events\n\nNo events defined.\n';

    let docs = '## Events\n\n';

    events.forEach((event) => {
        docs += `### \`${event.name}\`\n\n`;
        docs += '**Parameters:**\n';
        docs += generateEnhancedParameterDocs(event.inputs) + '\n\n';
    });

    return docs;
}

/**
 * Generates error documentation
 */
function generateErrorDocs(errors: ErrorInfo[]): string {
    if (errors.length === 0) return '## Errors\n\nNo custom errors defined.\n';

    let docs = '## Errors\n\n';

    errors.forEach((error) => {
        docs += `### \`${error.name}\`\n\n`;
        if (error.inputs.length > 0) {
            docs += '**Parameters:**\n';
            docs += generateEnhancedParameterDocs(error.inputs) + '\n\n';
        }
    });

    return docs;
}

/**
 * Generates complete contract documentation with enhanced context
 */
function generateContractDocs(contract: ContractInfo): string {
    let docs = `# ${contract.name} Contract\n\n`;

    // Add contract overview
    docs += `Complete reference for the ${contract.name} smart contract with ${contract.functions.length} functions, ${contract.events.length} events, and ${contract.errors.length} custom errors.\n\n`;

    // Add enhanced function documentation
    docs += generateEnhancedFunctionDocs(contract.functions);
    docs += generateEventDocs(contract.events);
    docs += generateErrorDocs(contract.errors);

    docs += '## Type Safety\n\n';
    docs += 'This contract is fully type-safe when used with the generated TypeScript types:\n\n';
    docs += '```typescript\n';
    docs += `import { ${contract.name.toLowerCase()}Abi } from '@layer-3/nitrolite';\n\n`;
    docs += `// Full type safety with autocomplete\n`;
    docs += `const result = await publicClient.readContract({\n`;
    docs += `  address: contractAddress,\n`;
    docs += `  abi: ${contract.name.toLowerCase()}Abi,\n`;
    docs += `  functionName: 'functionName', // ✅ Autocomplete available\n`;
    docs += `  args: [...], // ✅ Type-checked arguments\n`;
    docs += `});\n`;
    docs += '```\n\n';

    return docs;
}

/**
 * Generates SDK overview documentation
 */
function generateSDKOverview(contracts: ContractInfo[]): string {
    let docs = '# Nitrolite SDK Documentation\n\n';
    docs += '> **Auto-generated documentation with real usage examples**\n\n';
    docs +=
        'The Nitrolite SDK empowers developers to build high-performance, scalable web3 applications using state channels.\n\n';

    docs += '## Quick Start\n\n';
    docs += '```bash\n';
    docs += 'npm install @layer-3/nitrolite\n';
    docs += '```\n\n';

    docs += '```typescript\n';
    docs += "import { custodyAbi, NitroliteClient } from '@layer-3/nitrolite';\n\n";
    docs += '// Initialize client with full type safety\n';
    docs += 'const client = new NitroliteClient({...config});\n\n';
    docs += '// Deposit funds for state channels\n';
    docs += 'await client.deposit(tokenAddress, amount);\n\n';
    docs += '// Create a state channel\n';
    docs += 'const { channelId } = await client.createChannel({\n';
    docs += '  initialAllocationAmounts: [amount1, amount2]\n';
    docs += '});\n';
    docs += '```\n\n';

    docs += '## Available Contracts\n\n';
    contracts.forEach((contract) => {
        docs += `### ${contract.name}\n`;
        docs += `- **Functions:** ${contract.functions.length}\n`;
        docs += `- **Events:** ${contract.events.length}\n`;
        docs += `- **Errors:** ${contract.errors.length}\n`;
        docs += `- [View Details](./contracts/${contract.name}.md)\n\n`;
    });

    docs += '## Key Features\n\n';
    docs += '✅ **Auto-generated Types** - Always synchronized with contract changes\n';
    docs += '✅ **Real Usage Examples** - Extracted from actual codebase usage\n';
    docs += '✅ **Business Context** - Meaningful descriptions from JSDoc comments\n';
    docs += '✅ **Type Safety** - Full TypeScript support with autocomplete\n';
    docs += '✅ **Zero Maintenance** - Documentation updates automatically\n\n';

    return docs;
}

/**
 * Main function to generate all documentation
 */
function main() {
    console.log('Generating SDK documentation...');

    try {
        // Extract contract information
        const contracts = extractContractInfo();
        console.log(`Found ${contracts.length} contracts`);

        // Create docs directory structure
        const docsDir = join(__dirname, '../docs');
        const contractsDir = join(docsDir, 'contracts');

        // Ensure directories exist
        try {
            readdirSync(docsDir);
        } catch {
            require('fs').mkdirSync(docsDir, { recursive: true });
        }

        try {
            readdirSync(contractsDir);
        } catch {
            require('fs').mkdirSync(contractsDir, { recursive: true });
        }

        // Generate SDK overview
        const overviewDocs = generateSDKOverview(contracts);
        writeFileSync(join(docsDir, 'README.md'), overviewDocs);
        console.log('Generated SDK overview documentation');

        // Generate individual contract documentation
        contracts.forEach((contract) => {
            const contractDocs = generateContractDocs(contract);
            writeFileSync(join(contractsDir, `${contract.name}.md`), contractDocs);
            console.log(`Generated documentation for ${contract.name}`);
        });

        console.log(`Documentation generation complete! Generated docs for ${contracts.length} contracts.`);
        console.log(`Documentation available at: ${docsDir}`);
    } catch (error) {
        console.error('❌ Error generating documentation:', error);
        process.exit(1);
    }
}

if (require.main === module) {
    main();
}
