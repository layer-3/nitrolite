import { doc } from 'prettier';
import { printSeparatedList } from '../common/printer-helpers.js';
const { group } = doc.builders;
const contents = (node, path, print) => {
    var _a;
    return ((_a = node.components) === null || _a === void 0 ? void 0 : _a.length) === 1 && node.components[0].type === 'BinaryOperation'
        ? path.map(print, 'components')
        : printSeparatedList(path.map(print, 'components'));
};
export const TupleExpression = {
    print: ({ node, path, print }) => group([
        node.isArray ? '[' : '(',
        contents(node, path, print),
        node.isArray ? ']' : ')'
    ])
};
//# sourceMappingURL=TupleExpression.js.map