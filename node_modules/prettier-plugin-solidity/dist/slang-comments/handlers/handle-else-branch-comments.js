import { NonterminalKind } from '@nomicfoundation/slang/cst';
import { util } from 'prettier';
import addCollectionFirstComment from './add-collection-first-comment.js';
const { addLeadingComment } = util;
export default function handleElseBranchComments({ enclosingNode, followingNode, comment }) {
    if ((enclosingNode === null || enclosingNode === void 0 ? void 0 : enclosingNode.kind) !== NonterminalKind.ElseBranch || !followingNode) {
        return false;
    }
    if (followingNode === enclosingNode.body &&
        followingNode.kind === NonterminalKind.IfStatement) {
        if (followingNode.body.kind === NonterminalKind.Block) {
            addCollectionFirstComment(followingNode.body.statements, comment);
        }
        else {
            addLeadingComment(followingNode.body, comment);
        }
        return true;
    }
    return false;
}
//# sourceMappingURL=handle-else-branch-comments.js.map