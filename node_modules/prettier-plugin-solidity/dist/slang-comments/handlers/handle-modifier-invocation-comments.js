import { NonterminalKind } from '@nomicfoundation/slang/cst';
import { util } from 'prettier';
import { locEnd } from '../../slang-utils/loc.js';
import addCollectionFirstComment from './add-collection-first-comment.js';
const { addTrailingComment } = util;
export default function handleModifierInvocationComments({ text, precedingNode, enclosingNode, followingNode, comment }) {
    if ((enclosingNode === null || enclosingNode === void 0 ? void 0 : enclosingNode.kind) !== NonterminalKind.ModifierInvocation) {
        return false;
    }
    const nextCharacter = util.getNextNonSpaceNonCommentCharacter(text, locEnd(comment));
    // The last comments before the body.
    if ((precedingNode === null || precedingNode === void 0 ? void 0 : precedingNode.kind) === NonterminalKind.IdentifierPath &&
        nextCharacter === '(' &&
        (followingNode === null || followingNode === void 0 ? void 0 : followingNode.kind) === NonterminalKind.PositionalArgumentsDeclaration) {
        if (followingNode.arguments.items.length === 0) {
            addTrailingComment(enclosingNode, comment);
        }
        else {
            addCollectionFirstComment(followingNode.arguments, comment);
        }
        return true;
    }
    return false;
}
//# sourceMappingURL=handle-modifier-invocation-comments.js.map