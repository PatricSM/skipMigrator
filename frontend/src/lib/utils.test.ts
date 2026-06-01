import { describe, it, expect } from 'vitest'
import { cn } from './utils'

describe('cn (Tailwind class merger)', () => {
  it('joins truthy class names', () => {
    expect(cn('a', 'b', 'c')).toBe('a b c')
  })

  it('filters falsy values', () => {
    expect(cn('a', false, undefined, 'b', null, '', 'c')).toBe('a b c')
  })

  it('dedupes conflicting tailwind utilities (keeps the last)', () => {
    // tailwind-merge resolves padding conflicts → "px-4" wins over "px-2"
    expect(cn('px-2', 'px-4')).toBe('px-4')
  })

  it('flattens arrays and objects (clsx semantics)', () => {
    expect(cn(['a', 'b'], { c: true, d: false })).toBe('a b c')
  })

  it('returns empty string when given nothing actionable', () => {
    expect(cn(false, null, undefined)).toBe('')
  })
})
