// see: https://github.com/prettier/prettier/blob/main/src/language-js/loc.js
function getRange(index, node) {
    var _a;
    if (node.range) {
        return node.range[index];
    }
    if ((_a = node.expression) === null || _a === void 0 ? void 0 : _a.range) {
        return node.expression.range[index];
    }
    return null;
}
export default {
    locEnd: (node) => getRange(1, node),
    locStart: (node) => getRange(0, node)
};
//# sourceMappingURL=loc.js.map