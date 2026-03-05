# 🎯 TypeChain Automation for Nitrolite SDK

> Complete TypeScript type generation and automation system using Wagmi CLI + Foundry

## 🚀 What This Solves

**Before:** Manual ABI maintenance, potential desynchronization, no type safety  
**After:** Auto-generated types, always in sync, full type safety, zero maintenance

## ✨ Features

- ✅ **Auto-generated ABIs** - Always synchronized with contract changes
- ✅ **Full type safety** - Catch errors at compile time, not runtime
- ✅ **Viem integration** - Native support for Viem's type inference
- ✅ **Zero maintenance** - Contract changes automatically update TypeScript types
- ✅ **Complex types** - Handles structs, events, errors, and nested types
- ✅ **Build integration** - Automatic generation during build process

## 🛠️ Configuration

### 1. Wagmi CLI Configuration (`wagmi.config.ts`)

```typescript
import { defineConfig } from '@wagmi/cli';
import { foundry } from '@wagmi/cli/plugins';

export default defineConfig({
    out: 'src/generated.ts',
    contracts: [],
    plugins: [
        foundry({
            project: '../contract',
            include: [
                'Custody.sol/**',
                // Add other contracts as needed
            ],
            exclude: ['*.t.sol/**', '*.s.sol/**', 'forge-std/**', 'openzeppelin-contracts/**'],
        }),
    ],
});
```

### 2. Build Integration (`package.json`)

```json
{
    "scripts": {
        "codegen": "wagmi generate",
        "build:prod": "npm run codegen && tsc -p tsconfig.prod.json",
        "build:full": "npm run validate && npm run build && npm run docs && npm run docs:tutorials"
    }
}
```

### 3. Usage in Code

```typescript
import { custodyAbi } from '@layer-3/nitrolite';

// ✅ Full type safety and autocomplete
const result = await publicClient.readContract({
    address: CUSTODY_ADDRESS,
    abi: custodyAbi,
    functionName: 'getAccountInfo', // ✅ Auto-complete
    args: [userAddress, tokenAddress], // ✅ Type-checked
});
```

## 🔄 Development Workflow

```bash
# 1. Make contract changes
vim contract/src/Custody.sol

# 2. Build contracts
cd contract && forge build

# 3. Regenerate types
cd ../sdk && npm run codegen

# 4. Build SDK
npm run build
```

## 📋 Current Implementation Status

- ✅ **Wagmi CLI configured** - Basic setup complete
- ✅ **Build integration** - Codegen runs during build
- 🔄 **Migration in progress** - Moving from manual to generated ABIs
- 📝 **Team adoption** - Gradual rollout to team

### Generated Contracts

Currently generates types for:

- **`custodyAbi`** - Main custody contract
- **`adjudicatorAbi`** - Adjudicator contract
- **Additional contracts** - As configured in wagmi.config.ts

## 👥 Team Adoption Guide

### For Developers

1. **Run codegen**: `npm run codegen` to see generated types
2. **Check generated.ts**: Review the auto-generated ABIs
3. **Update imports**: Gradually switch from manual to generated ABIs
4. **Test thoroughly**: Ensure type safety works as expected

### For Contract Changes

1. **Update contracts** in `../contract` directory
2. **Build contracts**: `forge build`
3. **Regenerate types**: `npm run build` (includes codegen)
4. **Commit both**: Contract changes AND updated generated types

## 🚀 Benefits Over Manual ABIs

| Manual ABIs                  | Auto-Generated      |
| ---------------------------- | ------------------- |
| ❌ Manual sync required      | ✅ Automatic sync   |
| ❌ Risk of desynchronization | ✅ Always in sync   |
| ❌ No type safety            | ✅ Full type safety |
| ❌ Manual maintenance        | ✅ Zero maintenance |

## 🔮 Next Steps

1. **Expand contract coverage** - Add more contracts to wagmi.config.ts
2. **Complete migration** - Replace all manual ABIs with generated ones
3. **Add validation** - Implement comprehensive type checking
4. **CI/CD integration** - Automated validation in pipelines

---

**This provides exactly what TypeChain offers, tailored for Viem + Foundry!** 🎯
