import { TerminalKind } from '@nomicfoundation/slang/cst';
import { createKindCheckFunction } from './create-kind-check-function.js';
export const isBlockComment = createKindCheckFunction([
    TerminalKind.MultiLineComment,
    TerminalKind.MultiLineNatSpecComment
]);
export const isComment = createKindCheckFunction([
    TerminalKind.MultiLineComment,
    TerminalKind.MultiLineNatSpecComment,
    TerminalKind.SingleLineComment,
    TerminalKind.SingleLineNatSpecComment
]);
//# sourceMappingURL=is-comment.js.map