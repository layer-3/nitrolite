import { TerminalKind } from '@nomicfoundation/slang/cst';
import { CommentNode } from './CommentNode.js';
export class SingleLineNatSpecComment extends CommentNode {
    constructor(ast, offset) {
        super(ast, offset);
        this.kind = TerminalKind.SingleLineNatSpecComment;
        this.value = ast.unparse();
    }
    print() {
        return this.value.trimEnd();
    }
}
//# sourceMappingURL=SingleLineNatSpecComment.js.map