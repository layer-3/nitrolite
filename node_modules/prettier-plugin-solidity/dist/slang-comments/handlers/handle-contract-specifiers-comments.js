import { NonterminalKind } from '@nomicfoundation/slang/cst';
import { util } from 'prettier';
import addCollectionLastComment from './add-collection-last-comment.js';
const { addTrailingComment } = util;
export default function handleContractSpecifiersComments({ precedingNode, enclosingNode, comment }) {
    if ((enclosingNode === null || enclosingNode === void 0 ? void 0 : enclosingNode.kind) !== NonterminalKind.ContractSpecifiers) {
        return false;
    }
    if (precedingNode) {
        if (precedingNode.kind === NonterminalKind.InheritanceSpecifier) {
            addCollectionLastComment(precedingNode.types, comment);
            return true;
        }
        if (precedingNode.kind === NonterminalKind.StorageLayoutSpecifier) {
            addTrailingComment(precedingNode.expression, comment);
            return true;
        }
    }
    return false;
}
//# sourceMappingURL=handle-contract-specifiers-comments.js.map