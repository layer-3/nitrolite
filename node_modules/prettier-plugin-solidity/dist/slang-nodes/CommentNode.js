export class CommentNode {
    constructor(ast, offset) {
        this.loc = {
            start: offset,
            end: offset + ast.textLength.utf16
        };
    }
}
//# sourceMappingURL=CommentNode.js.map