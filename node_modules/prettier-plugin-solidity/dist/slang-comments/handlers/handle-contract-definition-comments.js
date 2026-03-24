import { NonterminalKind } from '@nomicfoundation/slang/cst';
import { util } from 'prettier';
import { locEnd } from '../../slang-utils/loc.js';
import addCollectionLastComment from './add-collection-last-comment.js';
const { addLeadingComment, addTrailingComment } = util;
export default function handleContractDefinitionComments({ text, precedingNode, enclosingNode, followingNode, comment }) {
    if ((enclosingNode === null || enclosingNode === void 0 ? void 0 : enclosingNode.kind) !== NonterminalKind.ContractDefinition) {
        return false;
    }
    const nextCharacter = util.getNextNonSpaceNonCommentCharacter(text, locEnd(comment));
    // Everything before the ContractSpecifiers is pushed onto the beginning of
    // the ContractDefinition.
    if ((followingNode === null || followingNode === void 0 ? void 0 : followingNode.kind) === NonterminalKind.ContractSpecifiers ||
        ((followingNode === null || followingNode === void 0 ? void 0 : followingNode.kind) === NonterminalKind.ContractMembers &&
            nextCharacter !== '{')) {
        addLeadingComment(enclosingNode, comment);
        return true;
    }
    // The comment is at the end of the body of the ContractDefinition.
    if ((precedingNode === null || precedingNode === void 0 ? void 0 : precedingNode.kind) === NonterminalKind.ContractMembers) {
        addCollectionLastComment(precedingNode, comment);
        return true;
    }
    // The last comments before the body.
    if (nextCharacter === '{') {
        if ((precedingNode === null || precedingNode === void 0 ? void 0 : precedingNode.kind) === NonterminalKind.ContractSpecifiers) {
            if (precedingNode.items.length === 0) {
                addTrailingComment(precedingNode, comment);
                return true;
            }
            const lastContractSpecifier = precedingNode.items[precedingNode.items.length - 1];
            // If the last ContractSpecifier's an InheritanceSpecifier, the comment
            // is appended to the last InheritanceType.
            if (lastContractSpecifier.kind === NonterminalKind.InheritanceSpecifier) {
                addCollectionLastComment(lastContractSpecifier.types, comment);
                return true;
            }
            if (lastContractSpecifier.kind === NonterminalKind.StorageLayoutSpecifier) {
                addTrailingComment(lastContractSpecifier.expression, comment);
                return true;
            }
        }
    }
    return false;
}
//# sourceMappingURL=handle-contract-definition-comments.js.map