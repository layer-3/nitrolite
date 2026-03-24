import { isComment } from '../slang-utils/is-comment.js';
export function printComment({ node: comment }) {
    if (isComment(comment)) {
        return comment.print();
    }
    throw new Error(`Not a comment: ${JSON.stringify(comment)}`);
}
//# sourceMappingURL=printer.js.map