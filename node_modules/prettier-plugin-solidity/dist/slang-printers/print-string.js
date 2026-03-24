import { util } from 'prettier';
const SINGLE_QUOTE = "'";
const DOUBLE_QUOTE = '"';
export function printString(rawContent, options) {
    const preferred = options.singleQuote ? SINGLE_QUOTE : DOUBLE_QUOTE;
    const alternate = preferred === SINGLE_QUOTE ? DOUBLE_QUOTE : SINGLE_QUOTE;
    let preferredQuoteCount = 0;
    let alternateQuoteCount = 0;
    for (const character of rawContent) {
        if (character === preferred) {
            preferredQuoteCount++;
        }
        else if (character === alternate) {
            alternateQuoteCount++;
        }
    }
    const enclosingQuote = preferredQuoteCount > alternateQuoteCount ? alternate : preferred;
    // It might sound unnecessary to use `makeString` even if the string already
    // is enclosed with `enclosingQuote`, but it isn't. The string could contain
    // unnecessary escapes (such as in `"\'"`). Always using `makeString` makes
    // sure that we consistently output the minimum amount of escaped quotes.
    return util.makeString(rawContent, enclosingQuote);
}
//# sourceMappingURL=print-string.js.map