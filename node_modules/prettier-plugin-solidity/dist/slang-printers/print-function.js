import { NonterminalKind } from '@nomicfoundation/slang/cst';
import { doc } from 'prettier';
import { joinExisting } from '../slang-utils/join-existing.js';
const { dedent, group, indent, line } = doc.builders;
export function printFunction(functionName, node, path, print) {
    const body = node.body;
    return group([
        functionName,
        path.call(print, 'parameters'),
        indent(group([
            joinExisting(line, [
                path.call(print, 'attributes'),
                path.call(print, 'returns')
            ]),
            body && body.kind === NonterminalKind.Block ? dedent(line) : ''
        ]))
    ]);
}
export function printFunctionWithBody(functionName, node, path, print) {
    return [
        printFunction(functionName, node, path, print),
        path.call(print, 'body')
    ];
}
//# sourceMappingURL=print-function.js.map