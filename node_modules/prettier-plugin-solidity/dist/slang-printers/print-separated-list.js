import { doc } from 'prettier';
import { printSeparatedItem } from './print-separated-item.js';
const { join, line } = doc.builders;
// This function will add an indentation to the `list` and separate it from the
// rest of the `doc` in most cases by a `softline`.
// the list itself will be printed with a separator that in most cases is a
// comma (,) and a `line`
export function printSeparatedList(list, { firstSeparator, separator = [',', line], lastSeparator, grouped } = {}) {
    return printSeparatedItem(join(separator, list), {
        firstSeparator,
        lastSeparator,
        grouped
    });
}
//# sourceMappingURL=print-separated-list.js.map