import { doc } from 'prettier';
const { group, hardline, indent, softline } = doc.builders;
// This function will add an indentation to the `item` and separate it from the
// rest of the `doc` in most cases by a `softline`.
export function printSeparatedItem(item, { firstSeparator = softline, lastSeparator = firstSeparator, grouped = firstSeparator !== hardline } = {}) {
    const document = [
        indent([firstSeparator, item]),
        lastSeparator
    ];
    return grouped ? group(document) : document;
}
//# sourceMappingURL=print-separated-item.js.map