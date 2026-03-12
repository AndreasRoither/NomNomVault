import { createComponent } from 'solid-js'
import { renderToString } from 'solid-js/web'
import { describe, expect, it } from 'vitest'

import { HomePage } from './HomePage'

describe('grocery home page', () => {
  it('renders a minimal hello screen', () => {
    const html = renderToString(() => createComponent(HomePage, {}))

    expect(html).toContain('Hello from Grocery')
  })
})
