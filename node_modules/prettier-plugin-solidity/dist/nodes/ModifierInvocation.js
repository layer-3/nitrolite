import { printComments, printSeparatedList } from '../common/printer-helpers.js';
const modifierArguments = (node, path, print, options) => {
    var _a;
    if (node.arguments) {
        // We always print parentheses at this stage because the parser already
        // stripped them in FunctionDefinitions that are not a constructor.
        return node.arguments.length > 0
            ? ['(', printSeparatedList(path.map(print, 'arguments')), ')']
            : '()';
    }
    if ((_a = node.comments) === null || _a === void 0 ? void 0 : _a.some((comment) => !comment.leading && !comment.trailing && !comment.printed)) {
        // We print parentheses here because the comment is supposed to be a block
        // comment inside empty parentheses.
        //    modifier(/* comment */)
        return ['(', printComments(node, path, options), ')'];
    }
    return '';
};
export const ModifierInvocation = {
    print: ({ node, path, print, options }) => [
        node.name,
        modifierArguments(node, path, print, options)
    ]
};
//# sourceMappingURL=ModifierInvocation.js.map