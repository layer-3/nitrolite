// @TODO: add support for assembly language specifier
import { printString } from '../common/util.js';
import { printSeparatedList } from '../common/printer-helpers.js';
export const InlineAssemblyStatement = {
    print: ({ node, path, print, options }) => {
        var _a;
        return [
            'assembly ',
            node.language ? `${printString(node.language, options)} ` : '',
            ((_a = node.flags) === null || _a === void 0 ? void 0 : _a.length) > 0
                ? [
                    '(',
                    printSeparatedList(node.flags.map((flag) => printString(flag, options))),
                    ') '
                ]
                : '',
            path.call(print, 'body')
        ];
    }
};
//# sourceMappingURL=InlineAssemblyStatement.js.map