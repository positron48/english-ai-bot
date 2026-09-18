import { expect, it } from 'vitest'
import { renderMarkdown } from './markdown'

it('removes executable HTML from model and course content', () => {
  const html = renderMarkdown('<img src="x" onerror="alert(1)"><script>alert(2)</script><a href="javascript:alert(3)">link</a>')
  const container = document.createElement('div'); container.innerHTML = html
  expect(container.querySelector('script')).toBeNull()
  expect(container.querySelector('img')?.hasAttribute('onerror')).toBe(false)
  expect(container.querySelector('a')?.hasAttribute('href')).toBe(false)
})
it('preserves ordinary Markdown and useful formatting', () => {
  const html = renderMarkdown('**word**\nnext\n\n[link](https://example.com)')
  expect(html).toContain('<strong>word</strong>')
  expect(html).toContain('<br>')
  expect(html).toContain('href="https://example.com"')
})
