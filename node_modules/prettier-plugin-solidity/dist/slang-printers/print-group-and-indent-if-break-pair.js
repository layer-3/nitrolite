import { doc } from 'prettier';
const { group, indentIfBreak } = doc.builders;
export function printGroupAndIndentIfBreakPair(groupDoc, indentIfBreakDoc) {
    const groupId = Symbol('Slang.GroupAndIndentIfBreakPair');
    return [
        group(groupDoc, { id: groupId }),
        indentIfBreak(indentIfBreakDoc, { groupId })
    ];
}
//# sourceMappingURL=print-group-and-indent-if-break-pair.js.map