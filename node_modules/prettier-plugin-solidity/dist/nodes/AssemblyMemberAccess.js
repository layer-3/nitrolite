export const AssemblyMemberAccess = {
    print: ({ path, print }) => [
        path.call(print, 'expression'),
        '.',
        path.call(print, 'memberName')
    ]
};
//# sourceMappingURL=AssemblyMemberAccess.js.map