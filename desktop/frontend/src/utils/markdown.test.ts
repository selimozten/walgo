import { describe, it, expect } from 'vitest'
import { sanitizeHtml, escapeHtml, renderMarkdown } from './markdown'

// =============================================================================
// escapeHtml — HTML entity escaping
// =============================================================================

describe('escapeHtml', () => {
  it('escapes ampersands', () => {
    expect(escapeHtml('a & b')).toBe('a &amp; b')
  })

  it('escapes less-than', () => {
    expect(escapeHtml('a < b')).toBe('a &lt; b')
  })

  it('escapes greater-than', () => {
    expect(escapeHtml('a > b')).toBe('a &gt; b')
  })

  it('escapes double quotes', () => {
    expect(escapeHtml('say "hello"')).toBe('say &quot;hello&quot;')
  })

  it('escapes all special characters together', () => {
    expect(escapeHtml('<div class="a">b & c</div>')).toBe(
      '&lt;div class=&quot;a&quot;&gt;b &amp; c&lt;/div&gt;'
    )
  })

  it('returns empty string unchanged', () => {
    expect(escapeHtml('')).toBe('')
  })

  it('returns plain text unchanged', () => {
    expect(escapeHtml('hello world')).toBe('hello world')
  })

  it('handles multiple ampersands', () => {
    expect(escapeHtml('a & b & c')).toBe('a &amp; b &amp; c')
  })

  it('handles already-escaped entities', () => {
    // Double-escaping is expected behavior
    expect(escapeHtml('&amp;')).toBe('&amp;amp;')
  })
})

// =============================================================================
// sanitizeHtml — XSS prevention
// =============================================================================

describe('sanitizeHtml', () => {
  it('removes script tags with content', () => {
    const input = '<p>Safe</p><script>alert("xss")</script><p>Also safe</p>'
    const result = sanitizeHtml(input)
    expect(result).not.toContain('<script')
    expect(result).not.toContain('alert')
    expect(result).toContain('Safe')
  })

  it('removes self-closing script tags', () => {
    const input = '<p>text</p><script src="evil.js"></script>'
    const result = sanitizeHtml(input)
    expect(result).not.toContain('script')
    expect(result).not.toContain('evil.js')
  })

  it('removes script tags with attributes', () => {
    const input = '<script type="text/javascript">document.cookie</script>'
    const result = sanitizeHtml(input)
    expect(result).not.toContain('document.cookie')
  })

  it('removes onclick handlers', () => {
    const input = '<div onclick="alert(1)">click me</div>'
    const result = sanitizeHtml(input)
    expect(result).not.toContain('onclick')
    expect(result).not.toContain('alert')
  })

  it('removes onmouseover handlers', () => {
    const input = '<a onmouseover="steal()">hover me</a>'
    const result = sanitizeHtml(input)
    expect(result).not.toContain('onmouseover')
    expect(result).not.toContain('steal')
  })

  it('removes onerror handlers', () => {
    const input = '<img onerror="alert(1)" src="x">'
    const result = sanitizeHtml(input)
    expect(result).not.toContain('onerror')
  })

  it('removes onload handlers', () => {
    const input = '<body onload="malicious()">'
    const result = sanitizeHtml(input)
    expect(result).not.toContain('onload')
  })

  it('neutralizes javascript: hrefs', () => {
    const input = '<a href="javascript:alert(1)">click</a>'
    const result = sanitizeHtml(input)
    expect(result).not.toContain('javascript:')
  })

  it('neutralizes javascript: in src', () => {
    const input = '<img src="javascript:alert(1)">'
    const result = sanitizeHtml(input)
    expect(result).not.toContain('javascript:')
  })

  it('preserves safe HTML', () => {
    const input = '<p>Hello <strong>world</strong></p>'
    expect(sanitizeHtml(input)).toBe(input)
  })

  it('preserves safe links', () => {
    const input = '<a href="https://example.com">link</a>'
    expect(sanitizeHtml(input)).toBe(input)
  })

  it('handles empty string', () => {
    expect(sanitizeHtml('')).toBe('')
  })

  it('removes multiple event handlers on same element', () => {
    const input = '<div onclick="a()" onmouseover="b()">text</div>'
    const result = sanitizeHtml(input)
    expect(result).not.toContain('onclick')
    expect(result).not.toContain('onmouseover')
  })

  it('handles case-insensitive script tags', () => {
    const input = '<SCRIPT>alert(1)</SCRIPT>'
    const result = sanitizeHtml(input)
    expect(result).not.toContain('alert')
  })

  it('handles script tags with newlines', () => {
    const input = '<script>\nalert(1)\n</script>'
    const result = sanitizeHtml(input)
    expect(result).not.toContain('alert')
  })
})

// =============================================================================
// renderMarkdown — Markdown to HTML conversion
// =============================================================================

describe('renderMarkdown', () => {
  // --- Heading conversion ---
  it('renders h1', () => {
    const result = renderMarkdown('# Hello')
    expect(result).toContain('<h1')
    expect(result).toContain('Hello')
  })

  it('renders h2', () => {
    const result = renderMarkdown('## Section')
    expect(result).toContain('<h2')
    expect(result).toContain('Section')
  })

  it('renders h3', () => {
    const result = renderMarkdown('### Subsection')
    expect(result).toContain('<h3')
    expect(result).toContain('Subsection')
  })

  // --- Inline formatting ---
  it('renders bold text', () => {
    const result = renderMarkdown('This is **bold** text')
    expect(result).toContain('<strong')
    expect(result).toContain('bold')
  })

  it('renders italic text', () => {
    const result = renderMarkdown('This is *italic* text')
    expect(result).toContain('<em')
    expect(result).toContain('italic')
  })

  it('renders bold and italic together', () => {
    const result = renderMarkdown('**bold** and *italic*')
    expect(result).toContain('<strong')
    expect(result).toContain('<em')
  })

  // --- Links ---
  it('renders links', () => {
    const result = renderMarkdown('[Click here](https://example.com)')
    expect(result).toContain('<a')
    expect(result).toContain('href="https://example.com"')
    expect(result).toContain('Click here')
  })

  it('renders links with text', () => {
    const result = renderMarkdown('Visit [my site](https://example.com) now')
    expect(result).toContain('<a')
    expect(result).toContain('my site')
  })

  // --- Images ---
  it('renders images', () => {
    const result = renderMarkdown('![Alt text](image.png)')
    expect(result).toContain('<img')
    expect(result).toContain('alt="Alt text"')
    expect(result).toContain('src="image.png"')
  })

  // --- Line breaks ---
  it('converts newlines to <br>', () => {
    const result = renderMarkdown('Line 1\nLine 2')
    expect(result).toContain('<br>')
  })

  // --- Plain text ---
  it('handles plain text', () => {
    const result = renderMarkdown('Just plain text')
    expect(result).toContain('Just plain text')
  })

  it('handles empty string', () => {
    expect(renderMarkdown('')).toBe('')
  })

  // --- XSS prevention through renderMarkdown ---
  it('escapes HTML in markdown content', () => {
    const result = renderMarkdown('<script>alert("xss")</script>')
    // The content should be escaped first by escapeHtml, then sanitized
    expect(result).not.toContain('<script>')
    expect(result).not.toMatch(/<script[^&]/)
  })

  it('escapes angle brackets in text', () => {
    const result = renderMarkdown('Use <div> tags')
    // < and > should be escaped to &lt; and &gt;
    expect(result).toContain('&lt;div&gt;')
  })

  it('handles markdown with HTML injection attempt', () => {
    const result = renderMarkdown('# Title <img onerror="alert(1)" src=x>')
    expect(result).not.toContain('onerror')
  })

  it('prevents javascript: in markdown links after rendering', () => {
    // Markdown link syntax — the escapeHtml will turn this into escaped text
    const result = renderMarkdown('[click](javascript:alert(1))')
    // After escaping, the parentheses content should not be a live javascript: URL
    // because escapeHtml escapes the quotes and angle brackets first
    expect(result).not.toMatch(/href="javascript:/)
  })

  // --- Complex content ---
  it('renders full markdown document', () => {
    const markdown = `# My Post

## Introduction

This is **important** content with *emphasis*.

Check out [this link](https://example.com) for more info.

### Details

More text here.`

    const result = renderMarkdown(markdown)
    expect(result).toContain('<h1')
    expect(result).toContain('<h2')
    expect(result).toContain('<h3')
    expect(result).toContain('<strong')
    expect(result).toContain('<em')
    expect(result).toContain('<a')
    expect(result).toContain('My Post')
    expect(result).toContain('Introduction')
    expect(result).toContain('Details')
  })

  it('handles multiple headings of same level', () => {
    const result = renderMarkdown('## First\n\n## Second\n\n## Third')
    const h2Count = (result.match(/<h2/g) || []).length
    expect(h2Count).toBe(3)
  })

  it('handles consecutive bold and italic', () => {
    const result = renderMarkdown('**bold1** **bold2** *italic1* *italic2*')
    const strongCount = (result.match(/<strong/g) || []).length
    const emCount = (result.match(/<em/g) || []).length
    expect(strongCount).toBe(2)
    expect(emCount).toBe(2)
  })
})
