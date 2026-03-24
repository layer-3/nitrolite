export const shouldGroupOrIndent = ({ type, operator }, matchers) => type !== 'BinaryOperation' ||
    matchers.some((matcher) => matcher.match(operator));
//# sourceMappingURL=should-group-or-indent.js.map