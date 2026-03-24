import { NonterminalKind } from '@nomicfoundation/slang/cst';
import addCollectionFirstComment from './add-collection-first-comment.js';
import addCollectionLastComment from './add-collection-last-comment.js';
export default function handleBlockComments({ precedingNode, enclosingNode, followingNode, comment }) {
    if ((enclosingNode === null || enclosingNode === void 0 ? void 0 : enclosingNode.kind) !== NonterminalKind.ParametersDeclaration) {
        return false;
    }
    if ((precedingNode === null || precedingNode === void 0 ? void 0 : precedingNode.kind) === NonterminalKind.Parameters) {
        addCollectionLastComment(precedingNode, comment);
        return true;
    }
    if ((followingNode === null || followingNode === void 0 ? void 0 : followingNode.kind) === NonterminalKind.Parameters) {
        addCollectionFirstComment(followingNode, comment);
        return true;
    }
    return false;
}
//# sourceMappingURL=handle-parameters-declaration-comments.js.map