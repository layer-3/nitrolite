import { TerminalKind } from '@nomicfoundation/slang/cst';
import { doc } from 'prettier';
import { isIndentableBlockComment } from '../slang-utils/is-indentable-block-comment.js';
import { printIndentableBlockComment } from '../slang-printers/print-indentable-block-comment.js';
import { CommentNode } from './CommentNode.js';
const { join, literalline } = doc.builders;
export class MultiLineNatSpecComment extends CommentNode {
    constructor(ast, offset) {
        super(ast, offset);
        this.kind = TerminalKind.MultiLineNatSpecComment;
        this.value = ast.unparse();
    }
    print() {
        if (isIndentableBlockComment(this)) {
            return printIndentableBlockComment(this);
        }
        return join(literalline, this.value.split('\n'));
    }
}
//# sourceMappingURL=MultiLineNatSpecComment.js.map