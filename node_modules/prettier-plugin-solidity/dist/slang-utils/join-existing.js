import { doc } from 'prettier';
const { join } = doc.builders;
export function joinExisting(sep, docs) {
    return join(sep, docs.filter(Boolean));
}
//# sourceMappingURL=join-existing.js.map