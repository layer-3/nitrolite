import { NonterminalKind } from '@nomicfoundation/slang/cst';
import addCollectionFirstComment from './add-collection-first-comment.js';
import addCollectionLastComment from './add-collection-last-comment.js';
export default function handlePositionalArgumentsDeclarationComments({ precedingNode, enclosingNode, followingNode, comment }) {
    if ((enclosingNode === null || enclosingNode === void 0 ? void 0 : enclosingNode.kind) !== NonterminalKind.PositionalArgumentsDeclaration) {
        return false;
    }
    if ((precedingNode === null || precedingNode === void 0 ? void 0 : precedingNode.kind) === NonterminalKind.PositionalArguments) {
        addCollectionLastComment(precedingNode, comment);
        return true;
    }
    if ((followingNode === null || followingNode === void 0 ? void 0 : followingNode.kind) === NonterminalKind.PositionalArguments) {
        addCollectionFirstComment(followingNode, comment);
        return true;
    }
    return false;
}
//# sourceMappingURL=handle-positional-arguments-declaration-comments.js.map