export function normalizeForRequest(source, fieldMap = {}) {
  const out = {}
  Object.keys(source || {}).forEach((key) => {
    const field = fieldMap[key] || {}
    const value = source[key]
    if (value === undefined) return
    if (typeof value === 'number' && !Number.isFinite(value)) {
      out[key] = null
      return
    }
    if (value === '' && field.inputType === 'select' && field.nullable === true) {
      out[key] = null
      return
    }
    out[key] = value
  })
  return out
}

export function stringifyPayload(source, fieldMap = {}) {
  return JSON.stringify(normalizeForRequest(source, fieldMap), null, 0)
}
