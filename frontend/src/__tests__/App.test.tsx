import { render, screen } from '@testing-library/react'
import App from '../App'

describe('App', () => {
  it('renders without errors', () => {
    render(<App />)
    expect(screen.getByText('Dojo CRM')).toBeInTheDocument()
  })
})
