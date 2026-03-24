import { NonterminalKind } from '@nomicfoundation/slang/cst';
import addCollectionFirstComment from './add-collection-first-comment.js';
import addCollectionLastComment from './add-collection-last-comment.js';
export default function handleSourceUnitMembersComments({ precedingNode, followingNode, comment }) {
    if (followingNode &&
        followingNode.kind === NonterminalKind.SourceUnitMembers) {
        addCollectionFirstComment(followingNode, comment);
        return true;
    }
    if (precedingNode &&
        precedingNode.kind === NonterminalKind.SourceUnitMembers) {
        addCollectionLastComment(precedingNode, comment);
        return true;
    }
    return false;
}
//# sourceMappingURL=handle-source-unit-members-comments.js.map