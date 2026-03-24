import { util } from 'prettier';
const { addDanglingComment, addTrailingComment } = util;
export default function addCollectionLastComment(node, comment) {
    if (node.items.length === 0) {
        addDanglingComment(node, comment, false);
    }
    else {
        addTrailingComment(node.items[node.items.length - 1], comment);
    }
}
//# sourceMappingURL=add-collection-last-comment.js.map