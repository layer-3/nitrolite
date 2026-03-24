export function locStart(node) {
    if (typeof node === 'string' || node === undefined)
        return -1;
    return node.loc.start;
}
export function locEnd(node) {
    if (typeof node === 'string' || node === undefined)
        return -1;
    return node.loc.end;
}
//# sourceMappingURL=loc.js.map