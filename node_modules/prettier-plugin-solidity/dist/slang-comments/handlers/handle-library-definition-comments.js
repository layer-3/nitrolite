import { NonterminalKind } from '@nomicfoundation/slang/cst';
import { util } from 'prettier';
import { locEnd } from '../../slang-utils/loc.js';
import addCollectionFirstComment from './add-collection-first-comment.js';
import addCollectionLastComment from './add-collection-last-comment.js';
export default function handleLibraryDefinitionComments({ text, precedingNode, enclosingNode, followingNode, comment }) {
    if ((enclosingNode === null || enclosingNode === void 0 ? void 0 : enclosingNode.kind) !== NonterminalKind.LibraryDefinition) {
        return false;
    }
    const nextCharacter = util.getNextNonSpaceNonCommentCharacter(text, locEnd(comment));
    // The comment is at the end of the body of the ContractDefinition.
    if ((precedingNode === null || precedingNode === void 0 ? void 0 : precedingNode.kind) === NonterminalKind.LibraryMembers) {
        addCollectionLastComment(precedingNode, comment);
        return true;
    }
    // The last comments before the body.
    if (nextCharacter === '{' &&
        (followingNode === null || followingNode === void 0 ? void 0 : followingNode.kind) === NonterminalKind.LibraryMembers) {
        addCollectionFirstComment(followingNode, comment);
        return true;
    }
    return false;
}
//# sourceMappingURL=handle-library-definition-comments.js.map