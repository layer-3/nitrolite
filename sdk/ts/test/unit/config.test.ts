import {
  APPLICATION_ID_QUERY_PARAM,
  appendApplicationIDQueryParam,
  type Config,
  withApplicationID,
} from '../../src/config';

describe('withApplicationID', () => {
  it('sets applicationID on the config', () => {
    const c: Config = { url: 'ws://host/' };
    withApplicationID('my-app')(c);
    expect(c.applicationID).toBe('my-app');
  });
});

describe('appendApplicationIDQueryParam', () => {
  it('returns the URL unchanged when applicationID is empty', () => {
    expect(appendApplicationIDQueryParam('ws://host/path')).toBe('ws://host/path');
    expect(appendApplicationIDQueryParam('ws://host/path', '')).toBe('ws://host/path');
  });

  it('adds app_id when no query string exists', () => {
    const out = appendApplicationIDQueryParam('ws://host/path', 'my-app');
    expect(out).toBe('ws://host/path?app_id=my-app');
  });

  it('adds app_id alongside an existing query string', () => {
    const out = appendApplicationIDQueryParam('ws://host/path?foo=bar', 'my-app');
    const parsed = new URL(out);
    expect(parsed.searchParams.get('foo')).toBe('bar');
    expect(parsed.searchParams.get(APPLICATION_ID_QUERY_PARAM)).toBe('my-app');
  });

  it('overwrites an existing app_id value', () => {
    const out = appendApplicationIDQueryParam('ws://host/path?app_id=old', 'new');
    expect(new URL(out).searchParams.get(APPLICATION_ID_QUERY_PARAM)).toBe('new');
  });

  it('url-encodes values with reserved characters', () => {
    const out = appendApplicationIDQueryParam('ws://host/', 'a b&c');
    expect(new URL(out).searchParams.get(APPLICATION_ID_QUERY_PARAM)).toBe('a b&c');
  });

  it('throws a descriptive error on an invalid URL', () => {
    expect(() => appendApplicationIDQueryParam('://', 'x')).toThrow(
      /cannot append app_id: invalid url/,
    );
  });
});
