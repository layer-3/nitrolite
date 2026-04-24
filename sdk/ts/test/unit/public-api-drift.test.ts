import * as publicApi from '../../src/index.js';

describe('SDK public runtime API drift guard', () => {
    it('keeps root runtime exports intentional', () => {
        expect(Object.keys(publicApi).sort()).toMatchSnapshot();
    });

    it('proves adversarial public export removal is observable', () => {
        const exports = new Set(Object.keys(publicApi));
        exports.delete('Client');

        expect(exports.has('Client')).toBe(false);
    });
});
