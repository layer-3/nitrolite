import { doc } from 'prettier';
import { printSeparatedList } from '../common/printer-helpers.js';
const { group, indentIfBreak, label, line, softline } = doc.builders;
const printObject = (path, print, options) => {
    const identifiers = path.map(print, 'identifiers');
    return [
        '{',
        printSeparatedList(path
            .map(print, 'arguments')
            .map((arg, index) => [identifiers[index], ': ', arg]), {
            firstSeparator: options.bracketSpacing ? line : softline,
            lastSeparator: [options.bracketSpacing ? line : softline, '})']
        })
    ];
};
const printArguments = (path, print) => printSeparatedList(path.map(print, 'arguments'), {
    lastSeparator: [softline, ')']
});
export const FunctionCall = {
    print: ({ node, path, print, options }) => {
        var _a, _b;
        let expressionDoc = path.call(print, 'expression');
        let argumentsDoc = ')';
        if (((_a = node.arguments) === null || _a === void 0 ? void 0 : _a.length) > 0) {
            if (((_b = node.identifiers) === null || _b === void 0 ? void 0 : _b.length) > 0) {
                argumentsDoc = printObject(path, print, options);
            }
            else {
                argumentsDoc = printArguments(path, print);
            }
        }
        // If we are at the end of a MemberAccessChain we should indent the
        // arguments accordingly.
        if (expressionDoc.label === 'MemberAccessChain') {
            expressionDoc = group(expressionDoc.contents, {
                id: Symbol('FunctionCall.expression')
            });
            argumentsDoc = indentIfBreak(argumentsDoc, {
                groupId: expressionDoc.id
            });
            // We wrap the expression in a label in case there is an IndexAccess or
            // a FunctionCall following this IndexAccess.
            return label('MemberAccessChain', [expressionDoc, '(', argumentsDoc]);
        }
        return [expressionDoc, '(', argumentsDoc].flat();
    }
};
//# sourceMappingURL=FunctionCall.js.map