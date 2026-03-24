import { doc } from 'prettier';
import { printComments } from '../common/printer-helpers.js';
const { hardline } = doc.builders;
export const ExpressionStatement = {
    print: ({ node, options, path, print }) => {
        var _a;
        const parts = [];
        const { parent } = path;
        if (parent.type === 'IfStatement') {
            if ((_a = node.comments) === null || _a === void 0 ? void 0 : _a.length) {
                const comments = printComments(node, path, options);
                if (comments === null || comments === void 0 ? void 0 : comments.length) {
                    parts.push(comments);
                    parts.push(hardline);
                }
            }
        }
        parts.push(path.call(print, 'expression'));
        parts.push(node.omitSemicolon ? '' : ';');
        return parts;
    }
};
//# sourceMappingURL=ExpressionStatement.js.map