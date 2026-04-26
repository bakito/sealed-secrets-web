import { describe, it, expect, beforeAll } from 'vitest'
import { createRequire } from 'module'
import { readFileSync } from 'fs'
import { fileURLToPath } from 'url'
import path from 'path'

// ---------------------------------------------------------------------------
// Bootstrap the YAML global from the vendored yaml.min.js (yaml.js library)
// ---------------------------------------------------------------------------
const __dirname = path.dirname(fileURLToPath(import.meta.url))
const require = createRequire(import.meta.url)

let YAML
beforeAll(() => {
  const src = readFileSync(path.join(__dirname, '../static/scripts/yaml.min.js'), 'utf8')
  const mod = { exports: {} }
  const windowStub = {}
  new Function('window', 'module', 'exports', 'require', src)(windowStub, mod, mod.exports, require) // eslint-disable-line no-new-func
  YAML = windowStub.YAML
})

// ---------------------------------------------------------------------------
// Inline of decodeSecretData from templates/index.html (kept in sync manually)
// ---------------------------------------------------------------------------
function decodeSecretData(content, format) {
  let parsed
  try {
    parsed = format === 'json'
      ? JSON.parse(content)
      : YAML.parse(content)
  } catch {
    return { content, error: 'Could not parse secret: invalid ' + format.toUpperCase() }
  }
  try {
    if (!parsed || !parsed.data || typeof parsed.data !== 'object') {
      return { content, error: null }
    }
    if (!parsed.stringData) {
      parsed.stringData = {}
    }
    for (const [key, val] of Object.entries(parsed.data)) {
      try {
        parsed.stringData[key] = val != null ? atob(val) : ''
      } catch {
        return { content, error: `Invalid base64 in key "${key}"` }
      }
    }
    delete parsed.data
    if (format === 'json') {
      return { content: JSON.stringify(parsed, null, 2), error: null }
    }
    const { stringData, ...rest } = parsed
    let yaml = '# loaded as plain text from secret\'s list\n' + YAML.stringify(rest)
    if (stringData && Object.keys(stringData).length > 0) {
      yaml += 'stringData:\n'
      for (const [key, val] of Object.entries(stringData)) {
        const crlfNormalized = val.replace(/\r\n/g, '\n')
        const hasStandaloneCR = crlfNormalized.includes('\r')
        const hasControlChars = /[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]/.test(val) || hasStandaloneCR
        if (typeof val === 'string' && !hasControlChars && crlfNormalized.includes('\n')) {
          const chomping = crlfNormalized.endsWith('\n') ? '|' : '|-'
          const body = crlfNormalized.endsWith('\n') ? crlfNormalized.slice(0, -1) : crlfNormalized
          yaml += `  ${key}: ${chomping}\n`
          for (const line of body.split('\n')) {
            yaml += `    ${line}\n`
          }
        } else {
          yaml += `  ${key}: ${YAML.stringify(val).trim()}\n`
        }
      }
    }
    return { content: yaml, error: null }
  } catch {
    return { content, error: 'Could not decode secret data' }
  }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
function b64(str) {
  return btoa(str)
}

function makeSecret(dataEntries) {
  return `apiVersion: v1\nkind: Secret\nmetadata:\n  name: test\n  namespace: default\ntype: Opaque\ndata:\n${Object.entries(dataEntries).map(([k, v]) => `  ${k}: ${b64(v)}`).join('\n')}\n`
}

// Convenience wrappers — most tests only care about the decoded content string
function decode(content, format = 'yaml') {
  return decodeSecretData(content, format).content
}

// Strip the leading comment line before round-trip parsing
function stripComment(yaml) {
  return yaml.replace(/^#.*\n/, '')
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe('decodeSecretData – passthrough cases', () => {
  it('returns content unchanged with no error when there is no data field', () => {
    const input = 'apiVersion: v1\nkind: Secret\nmetadata:\n  name: test\n'
    const result = decodeSecretData(input, 'yaml')
    expect(result.content).toBe(input)
    expect(result.error).toBeNull()
  })

  it('returns content unchanged with a parse error message on broken YAML', () => {
    // yaml.js throws on unclosed flow sequences/mappings
    const input = '{invalid: ['
    const result = decodeSecretData(input, 'yaml')
    expect(result.content).toBe(input)
    expect(result.error).toMatch(/invalid yaml/i)
  })

  it('returns content unchanged with a parse error message on broken JSON', () => {
    const input = '{ broken json'
    const result = decodeSecretData(input, 'json')
    expect(result.content).toBe(input)
    expect(result.error).toMatch(/invalid json/i)
  })
})

describe('decodeSecretData – single-line values (no line breaks)', () => {
  it('decodes a simple ASCII string', () => {
    const out = decode(makeSecret({ password: 'hunter2' }))
    expect(out).toContain('stringData:')
    expect(out).toContain('  password: hunter2')
    expect(out).not.toMatch(/^data:/m)
  })

  it('decodes a string with a tab (\\t) using double-quoted escaping', () => {
    const out = decode(makeSecret({ key: 'col1\tcol2' }))
    expect(out).toContain('stringData:')
    expect(out).toMatch(/key:.*\\t/)
    expect(out).not.toMatch(/key: \|/)
  })

  it('decodes a string with a null byte (\\0) using double-quoted escaping', () => {
    const out = decode(makeSecret({ key: 'val\x00end' }))
    expect(out).toContain('stringData:')
    expect(out).toMatch(/key:.*\\0/)
    expect(out).not.toMatch(/key: \|/)
  })

  it('decodes a string with \\x01–\\x1f control characters using double-quoted escaping', () => {
    const out = decode(makeSecret({ key: 'a\x01b\x1fc' }))
    expect(out).toContain('stringData:')
    expect(out).not.toMatch(/key: \|/)
  })

  it('decodes a string with DEL (\\x7f) using double-quoted escaping', () => {
    const out = decode(makeSecret({ key: 'a\x7fb' }))
    expect(out).toContain('stringData:')
    expect(out).not.toMatch(/key: \|/)
  })
})

describe('decodeSecretData – multiline values (\\n)', () => {
  it('uses block scalar | for a value with trailing newline', () => {
    const out = decode(makeSecret({ cfg: 'line1\nline2\n' }))
    expect(out).toContain('  cfg: |\n')
    expect(out).toContain('    line1\n')
    expect(out).toContain('    line2\n')
  })

  it('uses block scalar |- for a value without trailing newline', () => {
    const out = decode(makeSecret({ cfg: 'line1\nline2' }))
    expect(out).toContain('  cfg: |-\n')
    expect(out).toContain('    line1\n')
    expect(out).toContain('    line2\n')
  })

  it('round-trips correctly: YAML.parse of the output gives back the original value', () => {
    const val = 'key: value\nother: data\n'
    const out = decode(makeSecret({ 'values.yaml': val }))
    expect(YAML.parse(stripComment(out)).stringData['values.yaml']).toBe(val)
  })
})

describe('decodeSecretData – multiline values (\\r and \\r\\n)', () => {
  it('normalises \\r\\n (CRLF) line endings into a block scalar', () => {
    const out = decode(makeSecret({ cfg: 'line1\r\nline2\r\n' }))
    expect(out).toContain('  cfg: |\n')
    expect(out).toContain('    line1\n')
    expect(out).toContain('    line2\n')
    expect(out).not.toContain('\r')
  })

  it('falls back to double-quoted escaping for standalone \\r used as line endings (CR-only)', () => {
    const out = decode(makeSecret({ cfg: 'line1\rline2\r' }))
    expect(out).not.toMatch(/cfg: \|/)
    expect(out).not.toMatch(/cfg: \|-/)
    expect(out).toMatch(/cfg:.*\\r/)
  })

  it('falls back to double-quoted escaping for a \\r embedded mid-line (not a line ending)', () => {
    const out = decode(makeSecret({ key: 'He\rllo\nworld' }))
    expect(out).not.toMatch(/key: \|/)
    expect(out).toMatch(/key:.*\\r/)
  })

  it('round-trips \\r correctly via double-quoted escaping', () => {
    const val = 'line1\rline2\r'
    const out = decode(makeSecret({ cfg: val }))
    expect(YAML.parse(stripComment(out)).stringData.cfg).toBe(val)
  })

  it('CRLF round-trips correctly: output parses back to LF-normalised value', () => {
    const out = decode(makeSecret({ cfg: 'line1\r\nline2\r\n' }))
    // CRLF is normalised to LF in block scalar — this is the documented lossy step
    expect(YAML.parse(stripComment(out)).stringData.cfg).toBe('line1\nline2\n')
  })
})

describe('decodeSecretData – mixed: newline + control characters', () => {
  it('falls back to double-quoted escaping when value has \\n and \\0', () => {
    const out = decode(makeSecret({ key: 'line1\x00\nline2' }))
    expect(out).toContain('stringData:')
    expect(out).not.toMatch(/key: \|/)
    expect(out).not.toMatch(/key: \|-/)
  })

  it('falls back to double-quoted escaping when value has \\n and \\x01', () => {
    const out = decode(makeSecret({ key: 'col1\nval\x01' }))
    expect(out).not.toMatch(/key: \|/)
  })

  it('falls back to double-quoted escaping when value has \\n, \\r mid-line, and \\x01', () => {
    const out = decode(makeSecret({ key: 'col\t1\nval\x01\rend' }))
    expect(out).not.toMatch(/key: \|/)
  })

  it('allows \\t within a multiline block scalar (\\t is valid in block scalars)', () => {
    const out = decode(makeSecret({ cfg: 'col1\tcol2\nrow2\n' }))
    expect(out).toContain('  cfg: |\n')
    expect(out).toContain('    col1\tcol2\n')
  })
})

describe('decodeSecretData – multiple keys', () => {
  it('handles a mix of simple, multiline, and control-char values in one secret', () => {
    const secret = makeSecret({
      username: 'admin',
      password: 'p@ss',
      config: 'a: 1\nb: 2\n',
      binary: 'data\x00here',
    })
    const out = decode(secret)
    expect(out).toContain('  username: admin')
    expect(out).toContain('  password: p@ss')
    expect(out).toContain('  config: |\n')
    expect(out).toContain('    a: 1\n')
    expect(out).not.toMatch(/binary: \|/)
  })
})

describe('decodeSecretData – empty and blank values', () => {
  it('handles an empty string value — round-trips as empty string', () => {
    const out = decode(makeSecret({ key: '' }))
    expect(out).toContain('stringData:')
    expect(out).toMatch(/  key:/)
    expect(YAML.parse(stripComment(out)).stringData.key).toBe('')
  })

  it('handles a value that is only whitespace', () => {
    const out = decode(makeSecret({ key: '   ' }))
    expect(YAML.parse(stripComment(out)).stringData.key).toBe('   ')
  })

  it('handles a value with leading and trailing spaces', () => {
    const out = decode(makeSecret({ key: '  hello  ' }))
    expect(YAML.parse(stripComment(out)).stringData.key).toBe('  hello  ')
  })

  it('handles empty data object — no stringData block emitted', () => {
    const secret = 'apiVersion: v1\nkind: Secret\nmetadata:\n  name: test\n  namespace: default\ntype: Opaque\ndata: {}\n'
    const out = decode(secret)
    expect(out).not.toContain('stringData:')
  })
})

describe('decodeSecretData – YAML reserved words and ambiguous scalars', () => {
  it('quotes boolean-like values so they round-trip as strings', () => {
    for (const val of ['true', 'false', 'yes', 'no', 'on', 'off', 'True', 'False', 'YES', 'NO']) {
      const out = decode(makeSecret({ flag: val }))
      expect(YAML.parse(stripComment(out)).stringData.flag).toBe(val)
    }
  })

  it('quotes null-like values so they round-trip as strings', () => {
    for (const val of ['null', 'Null', 'NULL', '~']) {
      const out = decode(makeSecret({ key: val }))
      expect(YAML.parse(stripComment(out)).stringData.key).toBe(val)
    }
  })

  it('quotes numeric-looking values so they round-trip as strings', () => {
    for (const val of ['0', '42', '3.14', '-1', '1e10', '0x1F', '0o77', '1_000']) {
      const out = decode(makeSecret({ key: val }))
      expect(YAML.parse(stripComment(out)).stringData.key).toBe(val)
    }
  })
})

describe('decodeSecretData – YAML indicator characters', () => {
  it('quotes values starting with { (flow mapping indicator)', () => {
    const out = decode(makeSecret({ key: '{foo: bar}' }))
    expect(YAML.parse(stripComment(out)).stringData.key).toBe('{foo: bar}')
  })

  it('quotes values starting with [ (flow sequence indicator)', () => {
    const out = decode(makeSecret({ key: '[a, b]' }))
    expect(YAML.parse(stripComment(out)).stringData.key).toBe('[a, b]')
  })

  it('quotes values containing : followed by space (key-value separator)', () => {
    const out = decode(makeSecret({ key: 'host: localhost' }))
    expect(YAML.parse(stripComment(out)).stringData.key).toBe('host: localhost')
  })

  it('quotes values starting with # (comment character)', () => {
    const out = decode(makeSecret({ key: '# not a comment' }))
    expect(YAML.parse(stripComment(out)).stringData.key).toBe('# not a comment')
  })
})

describe('decodeSecretData – control chars \\x0b and \\x0c', () => {
  it('escapes VT (\\x0b) via double-quoted path', () => {
    const out = decode(makeSecret({ key: 'a\x0bb' }))
    expect(out).not.toMatch(/key: \|/)
    expect(YAML.parse(stripComment(out)).stringData.key).toBe('a\x0bb')
  })

  it('escapes FF (\\x0c) via double-quoted path', () => {
    const out = decode(makeSecret({ key: 'a\x0cb' }))
    expect(out).not.toMatch(/key: \|/)
    expect(YAML.parse(stripComment(out)).stringData.key).toBe('a\x0cb')
  })

  it('falls back to quoted when multiline value contains \\x0b', () => {
    const out = decode(makeSecret({ key: 'line1\x0b\nline2' }))
    expect(out).not.toMatch(/key: \|/)
  })
})

describe('decodeSecretData – mixed CRLF and standalone CR', () => {
  it('falls back to quoted when value mixes \\r\\n and standalone \\r', () => {
    const val = 'line1\r\nline2\rline3\n'
    const out = decode(makeSecret({ key: val }))
    expect(out).not.toMatch(/key: \|/)
    expect(YAML.parse(stripComment(out)).stringData.key).toBe(val)
  })
})

describe('decodeSecretData – invalid base64', () => {
  it('returns raw content and a descriptive error naming the bad key', () => {
    // yaml.js parses "hello!world" as a plain string; atob throws on "!"
    // Note: "!!!value" is parsed as null by yaml.js (YAML tag syntax) — must avoid leading "!"
    const secret = 'apiVersion: v1\nkind: Secret\nmetadata:\n  name: test\n  namespace: default\ntype: Opaque\ndata:\n  key: hello!world\n'
    const result = decodeSecretData(secret, 'yaml')
    expect(result.content).toBe(secret)
    expect(result.error).toMatch(/invalid base64/i)
    expect(result.error).toContain('"key"')
  })

  it('treats a null/empty base64 value as an empty string with no error', () => {
    // "  key: " (no value) → yaml.js returns null → null guard decodes to '' cleanly
    const secret = 'apiVersion: v1\nkind: Secret\nmetadata:\n  name: test\n  namespace: default\ntype: Opaque\ndata:\n  key:\n'
    const result = decodeSecretData(secret, 'yaml')
    expect(result.error).toBeNull()
    expect(YAML.parse(stripComment(result.content)).stringData.key).toBe('')
  })
})

describe('decodeSecretData – existing stringData merged with data', () => {
  it('data keys overwrite pre-existing stringData keys with the same name', () => {
    const secret = `apiVersion: v1\nkind: Secret\nmetadata:\n  name: test\n  namespace: default\ntype: Opaque\nstringData:\n  key: original\ndata:\n  key: ${b64('overwritten')}\n`
    const out = decode(secret)
    expect(YAML.parse(stripComment(out)).stringData.key).toBe('overwritten')
  })

  it('pre-existing stringData keys not present in data are preserved', () => {
    const secret = `apiVersion: v1\nkind: Secret\nmetadata:\n  name: test\n  namespace: default\ntype: Opaque\nstringData:\n  existing: kept\ndata:\n  new: ${b64('added')}\n`
    const out = decode(secret)
    const reparsed = YAML.parse(stripComment(out))
    expect(reparsed.stringData.existing).toBe('kept')
    expect(reparsed.stringData.new).toBe('added')
  })
})

describe('decodeSecretData – nested content with literal \\n sequences', () => {
  it('preserves literal \\n (two chars) inside a multiline block scalar unchanged', () => {
    // The value is a YAML snippet that itself contains a quoted string with \n escape.
    // The outer block scalar must not re-interpret those two chars as a newline.
    const val = 'message: "Hello\\nWorld"\nother: value\n'
    const out = decode(makeSecret({ 'config.yaml': val }))
    expect(out).toContain('  config.yaml: |\n')
    expect(out).toContain('    message: "Hello\\nWorld"')
    expect(YAML.parse(stripComment(out)).stringData['config.yaml']).toBe(val)
  })

  it('preserves literal \\r (two chars) inside a multiline block scalar unchanged', () => {
    const val = 'line: "value\\rhere"\nother: data\n'
    const out = decode(makeSecret({ key: val }))
    expect(YAML.parse(stripComment(out)).stringData.key).toBe(val)
  })
})

describe('decodeSecretData – JSON format', () => {
  it('decodes base64 values, moves data to stringData, and returns valid JSON', () => {
    const input = JSON.stringify({
      apiVersion: 'v1',
      kind: 'Secret',
      metadata: { name: 'test', namespace: 'default' },
      type: 'Opaque',
      data: { password: b64('secret') },
    })
    const out = decode(input, 'json')
    const parsed = JSON.parse(out)
    expect(parsed.stringData.password).toBe('secret')
    expect(parsed.data).toBeUndefined()
  })
})
