import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import SupabaseChecklist from './SupabaseChecklist'

describe('SupabaseChecklist', () => {
  const ref = 'abcd1234efgh5678ijkl'

  it('renders the four panel deep-links with the project ref baked in', () => {
    render(<SupabaseChecklist projectRef={ref} />)
    const links = screen.getAllByRole('link')
    const hrefs = links.map((a) => a.getAttribute('href'))
    expect(hrefs).toEqual(
      expect.arrayContaining([
        expect.stringContaining(`/dashboard/project/${ref}/auth/url-configuration`),
        expect.stringContaining(`/dashboard/project/${ref}/auth/providers`),
        expect.stringContaining(`/dashboard/project/${ref}/settings/api`),
        expect.stringContaining(`/dashboard/project/${ref}`),
      ]),
    )
  })

  it('shows the plug-and-play green callout', () => {
    render(<SupabaseChecklist projectRef={ref} />)
    expect(screen.getByText(/Não precisa mudar nada se você usar o mesmo Supabase/)).toBeInTheDocument()
  })

  it('embeds the gen types command with the actual project ref', () => {
    render(<SupabaseChecklist projectRef={ref} />)
    expect(
      screen.getByText((text) => text.includes(`--project-id ${ref}`)),
    ).toBeInTheDocument()
  })

  it('truncates the project ref in the subtitle', () => {
    render(<SupabaseChecklist projectRef={ref} />)
    // The full ref appears in URLs/code, but the subtitle uses the first 8 chars
    expect(screen.getByText(/abcd1234…/)).toBeInTheDocument()
  })
})
