const stateMutability = (node) => { var _a; return ((_a = node.stateMutability) === null || _a === void 0 ? void 0 : _a.length) > 0 ? [' ', node.stateMutability] : ''; };
export const ElementaryTypeName = {
    print: ({ node }) => [node.name, stateMutability(node)]
};
//# sourceMappingURL=ElementaryTypeName.js.map