import { doc } from 'prettier';
const { hardline, join } = doc.builders;
export function printIndentableBlockComment(comment) {
    const lines = comment.value.split('\n');
    return join(hardline, lines.map((line, index) => index === 0
        ? line.trimEnd()
        : ` ${index < lines.length - 1 ? line.trim() : line.trimStart()}`));
}
//# sourceMappingURL=print-indentable-block-comment.js.map