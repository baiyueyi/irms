import { describe, expect, it } from 'vitest'
import { normalizeForRequest, stringifyPayload } from '../crudPayload'

describe('crudPayload save action', () => {
  it('keeps valid JSON structure', () => {
    const json = stringifyPayload({ name: 'host-a', location_id: 10 }, { location_id: { inputType: 'select', nullable: true } })
    expect(JSON.parse(json)).toEqual({ name: 'host-a', location_id: 10 })
  })

  it('handles special characters safely', () => {
    const json = stringifyPayload(
      { description: 'a"b\\c\n中文', location_id: '' },
      { location_id: { inputType: 'select', nullable: true } }
    )
    expect(JSON.parse(json)).toEqual({ description: 'a"b\\c\n中文', location_id: null })
  })

  it('keeps empty string for non-nullable select', () => {
    const data = normalizeForRequest({ cloud_vendor: '' }, { cloud_vendor: { inputType: 'select', nullable: false } })
    expect(data).toEqual({ cloud_vendor: '' })
  })

  it('throws on invalid circular structure', () => {
    const source = { name: 'x' }
    source.self = source
    expect(() => stringifyPayload(source, {})).toThrow()
  })
})
