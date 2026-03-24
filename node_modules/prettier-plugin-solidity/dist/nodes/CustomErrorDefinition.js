import { printSeparatedList } from '../common/printer-helpers.js';
const parameters = (node, path, print) => {
    var _a;
    return ((_a = node.parameters) === null || _a === void 0 ? void 0 : _a.length) > 0
        ? printSeparatedList(path.map(print, 'parameters'))
        : '';
};
export const CustomErrorDefinition = {
    print: ({ node, path, print }) => [
        'error ',
        node.name,
        '(',
        parameters(node, path, print),
        ');'
    ]
};
//# sourceMappingURL=CustomErrorDefinition.js.map