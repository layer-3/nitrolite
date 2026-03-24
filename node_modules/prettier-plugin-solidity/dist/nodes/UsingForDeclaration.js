import { doc } from 'prettier';
import { printSeparatedList } from '../common/printer-helpers.js';
const { line, softline } = doc.builders;
export const UsingForDeclaration = {
    print: ({ node, path, print, options }) => {
        var _a;
        return [
            'using ',
            ((_a = node.functions) === null || _a === void 0 ? void 0 : _a.length)
                ? [
                    '{',
                    printSeparatedList(node.functions.map((functionName, i) => node.operators[i]
                        ? [functionName, ' as ', node.operators[i]]
                        : functionName), {
                        firstSeparator: options.bracketSpacing ? line : softline
                    }),
                    '}'
                ]
                : node.libraryName,
            ' for ',
            node.typeName ? path.call(print, 'typeName') : '*',
            node.isGlobal ? ' global;' : ';'
        ];
    }
};
//# sourceMappingURL=UsingForDeclaration.js.map