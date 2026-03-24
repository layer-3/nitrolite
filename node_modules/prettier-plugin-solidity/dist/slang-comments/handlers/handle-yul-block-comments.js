import { NonterminalKind } from '@nomicfoundation/slang/cst';
import { util } from 'prettier';
const { addLeadingComment, addDanglingComment } = util;
export default function handleYulBlockComments({ precedingNode, enclosingNode, followingNode, comment }) {
    if ((enclosingNode === null || enclosingNode === void 0 ? void 0 : enclosingNode.kind) !== NonterminalKind.YulBlock) {
        return false;
    }
    if ((precedingNode === null || precedingNode === void 0 ? void 0 : precedingNode.kind) === NonterminalKind.YulStatements &&
        precedingNode.items.length === 0) {
        addDanglingComment(precedingNode, comment, false);
        return true;
    }
    if ((followingNode === null || followingNode === void 0 ? void 0 : followingNode.kind) === NonterminalKind.YulStatements &&
        followingNode.items.length > 0) {
        addLeadingComment(followingNode.items[0], comment);
        return true;
    }
    return false;
}
//# sourceMappingURL=handle-yul-block-comments.js.map