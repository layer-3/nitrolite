import { doc } from 'prettier';
const { group, indent, line } = doc.builders;
export function printIndentedGroupOrSpacedDocument(document, shouldGroup = true, groupOptions = {}) {
    return shouldGroup
        ? group(indent([line, document]), groupOptions)
        : [' ', document];
}
//# sourceMappingURL=print-indented-group-or-spaced-document.js.map