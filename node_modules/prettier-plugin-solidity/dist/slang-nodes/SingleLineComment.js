import { TerminalKind } from '@nomicfoundation/slang/cst';
import { CommentNode } from './CommentNode.js';
export class SingleLineComment extends CommentNode {
    constructor(ast, offset) {
        super(ast, offset);
        this.kind = TerminalKind.SingleLineComment;
        this.value = ast.unparse();
    }
    print() {
        return this.value.trimEnd();
    }
}
//# sourceMappingURL=SingleLineComment.js.map