export const EmitStatement = {
    print: ({ path, print }) => ['emit ', path.call(print, 'eventCall'), ';']
};
//# sourceMappingURL=EmitStatement.js.map