import { NonterminalKind } from '@nomicfoundation/slang/cst';
export function sortContractSpecifiers({ kind: aKind }, { kind: bKind }) {
    // OverrideSpecifiers before ModifierInvocation
    if (aKind === NonterminalKind.InheritanceSpecifier &&
        bKind === NonterminalKind.StorageLayoutSpecifier)
        return -1;
    if (bKind === NonterminalKind.InheritanceSpecifier &&
        aKind === NonterminalKind.StorageLayoutSpecifier)
        return 1;
    return 0;
}
//# sourceMappingURL=sort-contract-specifiers.js.map