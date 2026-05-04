/**
 * Decodes the .data field of a Kubernetes Secret into .stringData.
 * @param {string} content The Secret content (YAML or JSON).
 * @param {string} format 'yaml' or 'json'.
 * @returns {{content: string, error: string|null}}
 */
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
