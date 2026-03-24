import { printSeparatedList } from '../common/printer-helpers.js';
const printArguments = (node, path, print) => {
    var _a;
    return ((_a = node.arguments) === null || _a === void 0 ? void 0 : _a.length)
        ? ['(', printSeparatedList(path.map(print, 'arguments')), ')']
        : '';
};
export const InheritanceSpecifier = {
    print: ({ node, path, print }) => [
        path.call(print, 'baseName'),
        printArguments(node, path, print)
    ]
};
//# sourceMappingURL=InheritanceSpecifier.js.map