import { NonterminalKind } from '@nomicfoundation/slang/cst';
import { util } from 'prettier';
const { addLeadingComment } = util;
export default function handleStorageLayoutSpecifierComments({ enclosingNode, followingNode, comment }) {
    if ((enclosingNode === null || enclosingNode === void 0 ? void 0 : enclosingNode.kind) !== NonterminalKind.StorageLayoutSpecifier) {
        return false;
    }
    if ((followingNode === null || followingNode === void 0 ? void 0 : followingNode.kind) === NonterminalKind.Expression) {
        addLeadingComment(followingNode, comment);
        return true;
    }
    return false;
}
//# sourceMappingURL=handle-storage-layout-specifier-comments.js.map