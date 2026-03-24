export const AssemblyIf = {
    print: ({ path, print }) => [
        'if ',
        path.call(print, 'condition'),
        ' ',
        path.call(print, 'body')
    ]
};
//# sourceMappingURL=AssemblyIf.js.map