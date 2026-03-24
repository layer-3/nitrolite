import { NonterminalKind } from '@nomicfoundation/slang/cst';
import { util } from 'prettier';
import { locEnd } from '../../slang-utils/loc.js';
import addCollectionFirstComment from './add-collection-first-comment.js';
import addCollectionLastComment from './add-collection-last-comment.js';
export default function handleStructComments({ text, precedingNode, enclosingNode, followingNode, comment }) {
    if ((enclosingNode === null || enclosingNode === void 0 ? void 0 : enclosingNode.kind) !== NonterminalKind.StructDefinition) {
        return false;
    }
    const nextCharacter = util.getNextNonSpaceNonCommentCharacter(text, locEnd(comment));
    if ((precedingNode === null || precedingNode === void 0 ? void 0 : precedingNode.kind) === NonterminalKind.StructMembers &&
        nextCharacter === '}') {
        addCollectionLastComment(precedingNode, comment);
        return true;
    }
    if ((followingNode === null || followingNode === void 0 ? void 0 : followingNode.kind) === NonterminalKind.StructMembers) {
        addCollectionFirstComment(followingNode, comment);
        return true;
    }
    return false;
}
//# sourceMappingURL=handle-struct-comments.js.map