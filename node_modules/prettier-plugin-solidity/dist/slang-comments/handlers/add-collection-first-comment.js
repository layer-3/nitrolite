import { util } from 'prettier';
const { addDanglingComment, addLeadingComment } = util;
export default function addCollectionFirstComment(node, comment) {
    if (node.items.length === 0) {
        addDanglingComment(node, comment, false);
    }
    else {
        addLeadingComment(node.items[0], comment);
    }
}
//# sourceMappingURL=add-collection-first-comment.js.map