import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { cn, formatFileSize, truncateText, formatRelativeTime, debounce, generateId } from './helpers'

// =============================================================================
// cn — Tailwind class merging
// =============================================================================

describe('cn', () => {
  it('merges simple classes', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('handles conditional classes', () => {
    expect(cn('base', false && 'hidden', true && 'visible')).toBe('base visible')
  })

  it('deduplicates tailwind classes', () => {
    // twMerge should resolve conflicting tailwind utilities
    const result = cn('p-4', 'p-2')
    expect(result).toBe('p-2')
  })

  it('handles empty inputs', () => {
    expect(cn()).toBe('')
  })

  it('handles undefined and null', () => {
    expect(cn('base', undefined, null, 'end')).toBe('base end')
  })

  it('handles arrays', () => {
    expect(cn(['foo', 'bar'])).toBe('foo bar')
  })

  it('merges conflicting tailwind colors', () => {
    const result = cn('text-red-500', 'text-blue-500')
    expect(result).toBe('text-blue-500')
  })

  it('keeps non-conflicting tailwind classes', () => {
    const result = cn('bg-red-500', 'text-white', 'p-4')
    expect(result).toContain('bg-red-500')
    expect(result).toContain('text-white')
    expect(result).toContain('p-4')
  })
})

// =============================================================================
// formatFileSize — Human-readable file sizes
// =============================================================================

describe('formatFileSize', () => {
  it('returns "0 Bytes" for 0', () => {
    expect(formatFileSize(0)).toBe('0 Bytes')
  })

  it('formats bytes (< 1KB)', () => {
    expect(formatFileSize(500)).toBe('500 Bytes')
  })

  it('formats exactly 1 byte', () => {
    expect(formatFileSize(1)).toBe('1 Bytes')
  })

  it('formats kilobytes', () => {
    expect(formatFileSize(1024)).toBe('1 KB')
  })

  it('formats kilobytes with decimals', () => {
    expect(formatFileSize(1536)).toBe('1.5 KB')
  })

  it('formats megabytes', () => {
    expect(formatFileSize(1048576)).toBe('1 MB')
  })

  it('formats megabytes with decimals', () => {
    const result = formatFileSize(1572864) // 1.5 MB
    expect(result).toBe('1.5 MB')
  })

  it('formats gigabytes', () => {
    expect(formatFileSize(1073741824)).toBe('1 GB')
  })

  it('formats large megabyte values', () => {
    const result = formatFileSize(104857600) // 100 MB
    expect(result).toBe('100 MB')
  })

  it('rounds to 2 decimal places', () => {
    const result = formatFileSize(1024 * 1024 * 2.567)
    // Should be approximately "2.57 MB"
    expect(result).toMatch(/^2\.5[67] MB$/)
  })
})

// =============================================================================
// truncateText — Text truncation with ellipsis
// =============================================================================

describe('truncateText', () => {
  it('returns short text unchanged', () => {
    expect(truncateText('hello', 10)).toBe('hello')
  })

  it('returns text at exact limit unchanged', () => {
    expect(truncateText('hello', 5)).toBe('hello')
  })

  it('truncates long text with ellipsis', () => {
    expect(truncateText('hello world', 5)).toBe('hello...')
  })

  it('handles empty string', () => {
    expect(truncateText('', 10)).toBe('')
  })

  it('handles maxLength of 0', () => {
    expect(truncateText('hello', 0)).toBe('...')
  })

  it('handles single character', () => {
    expect(truncateText('a', 1)).toBe('a')
  })

  it('truncates long sentence', () => {
    const text = 'This is a very long sentence that should be truncated'
    const result = truncateText(text, 20)
    expect(result).toBe('This is a very long ...')
    expect(result.length).toBe(23) // 20 + '...'
  })

  it('handles unicode text', () => {
    const text = 'Hello world'
    const result = truncateText(text, 5)
    expect(result).toBe('Hello...')
  })
})

// =============================================================================
// formatRelativeTime — Relative time formatting
// =============================================================================

describe('formatRelativeTime', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-02-07T12:00:00Z'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('returns "just now" for less than 60 seconds ago', () => {
    const date = new Date('2026-02-07T11:59:30Z') // 30 seconds ago
    expect(formatRelativeTime(date)).toBe('just now')
  })

  it('formats minutes ago', () => {
    const date = new Date('2026-02-07T11:55:00Z') // 5 minutes ago
    expect(formatRelativeTime(date)).toBe('5m ago')
  })

  it('formats 1 minute ago', () => {
    const date = new Date('2026-02-07T11:59:00Z') // 1 minute ago
    expect(formatRelativeTime(date)).toBe('1m ago')
  })

  it('formats 59 minutes ago', () => {
    const date = new Date('2026-02-07T11:01:00Z') // 59 minutes ago
    expect(formatRelativeTime(date)).toBe('59m ago')
  })

  it('formats hours ago', () => {
    const date = new Date('2026-02-07T09:00:00Z') // 3 hours ago
    expect(formatRelativeTime(date)).toBe('3h ago')
  })

  it('formats 1 hour ago', () => {
    const date = new Date('2026-02-07T11:00:00Z') // 1 hour ago
    expect(formatRelativeTime(date)).toBe('1h ago')
  })

  it('formats days ago', () => {
    const date = new Date('2026-02-05T12:00:00Z') // 2 days ago
    expect(formatRelativeTime(date)).toBe('2d ago')
  })

  it('formats 6 days ago', () => {
    const date = new Date('2026-02-01T12:00:00Z') // 6 days ago
    expect(formatRelativeTime(date)).toBe('6d ago')
  })

  it('returns formatted date for 7+ days ago', () => {
    const date = new Date('2026-01-20T12:00:00Z') // 18 days ago
    const result = formatRelativeTime(date)
    // Should be a locale date string, not a relative time
    expect(result).not.toContain('ago')
    expect(result).not.toBe('just now')
  })

  it('accepts string dates', () => {
    const result = formatRelativeTime('2026-02-07T11:58:00Z') // 2 minutes ago
    expect(result).toBe('2m ago')
  })

  it('accepts ISO date strings', () => {
    const result = formatRelativeTime('2026-02-07T11:00:00.000Z')
    expect(result).toBe('1h ago')
  })
})

// =============================================================================
// debounce — Delayed function execution
// =============================================================================

describe('debounce', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('delays function execution', () => {
    const fn = vi.fn()
    const debounced = debounce(fn, 100)

    debounced()
    expect(fn).not.toHaveBeenCalled()

    vi.advanceTimersByTime(100)
    expect(fn).toHaveBeenCalledTimes(1)
  })

  it('resets timer on subsequent calls', () => {
    const fn = vi.fn()
    const debounced = debounce(fn, 100)

    debounced()
    vi.advanceTimersByTime(50)
    debounced() // Reset timer
    vi.advanceTimersByTime(50)

    // Should NOT have been called yet (only 50ms since last call)
    expect(fn).not.toHaveBeenCalled()

    vi.advanceTimersByTime(50)
    expect(fn).toHaveBeenCalledTimes(1)
  })

  it('passes arguments to the debounced function', () => {
    const fn = vi.fn()
    const debounced = debounce(fn, 100)

    debounced('arg1', 'arg2')
    vi.advanceTimersByTime(100)

    expect(fn).toHaveBeenCalledWith('arg1', 'arg2')
  })

  it('uses latest arguments when called multiple times', () => {
    const fn = vi.fn()
    const debounced = debounce(fn, 100)

    debounced('first')
    debounced('second')
    debounced('third')
    vi.advanceTimersByTime(100)

    expect(fn).toHaveBeenCalledTimes(1)
    expect(fn).toHaveBeenCalledWith('third')
  })

  it('can be called again after debounce completes', () => {
    const fn = vi.fn()
    const debounced = debounce(fn, 100)

    debounced('call1')
    vi.advanceTimersByTime(100)
    expect(fn).toHaveBeenCalledTimes(1)

    debounced('call2')
    vi.advanceTimersByTime(100)
    expect(fn).toHaveBeenCalledTimes(2)
    expect(fn).toHaveBeenLastCalledWith('call2')
  })

  it('handles zero delay', () => {
    const fn = vi.fn()
    const debounced = debounce(fn, 0)

    debounced()
    vi.advanceTimersByTime(0)
    expect(fn).toHaveBeenCalledTimes(1)
  })

  it('handles rapid-fire calls', () => {
    const fn = vi.fn()
    const debounced = debounce(fn, 100)

    for (let i = 0; i < 100; i++) {
      debounced(i)
    }

    vi.advanceTimersByTime(100)
    expect(fn).toHaveBeenCalledTimes(1)
    expect(fn).toHaveBeenCalledWith(99) // Last call wins
  })
})

// =============================================================================
// generateId — Unique ID generation
// =============================================================================

describe('generateId', () => {
  it('returns a non-empty string', () => {
    const id = generateId()
    expect(id).toBeTruthy()
    expect(typeof id).toBe('string')
  })

  it('contains a timestamp component', () => {
    const before = Date.now()
    const id = generateId()
    const after = Date.now()

    const timestamp = parseInt(id.split('-')[0], 10)
    expect(timestamp).toBeGreaterThanOrEqual(before)
    expect(timestamp).toBeLessThanOrEqual(after)
  })

  it('contains a random component after the dash', () => {
    const id = generateId()
    const parts = id.split('-')
    expect(parts.length).toBe(2)
    expect(parts[1].length).toBeGreaterThan(0)
  })

  it('generates unique IDs', () => {
    const ids = new Set<string>()
    for (let i = 0; i < 1000; i++) {
      ids.add(generateId())
    }
    // With timestamp + random, all 1000 should be unique
    expect(ids.size).toBe(1000)
  })

  it('matches expected format (timestamp-random)', () => {
    const id = generateId()
    // Should be: digits, dash, alphanumeric
    expect(id).toMatch(/^\d+-[a-z0-9]+$/)
  })
})
