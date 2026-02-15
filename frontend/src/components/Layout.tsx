import { Link, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'

interface NavItem {
  label: string
  to: string
  roles: Array<'admin' | 'instructor' | 'user'>
}

const navItems: NavItem[] = [
  { label: 'Dashboard', to: '/', roles: ['admin', 'instructor', 'user'] },
  { label: 'Members', to: '/members', roles: ['admin', 'instructor'] },
  { label: 'Classes', to: '/classes', roles: ['admin', 'instructor'] },
  { label: 'Attendance', to: '/attendance', roles: ['admin', 'instructor'] },
  { label: 'Payments', to: '/payments', roles: ['admin'] },
  { label: 'My Attendance', to: '/my-attendance', roles: ['user'] },
  { label: 'My Payments', to: '/my-payments', roles: ['user'] },
]

export function Layout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  const visibleItems = navItems.filter(
    (item) => user && item.roles.includes(user.role),
  )

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="layout">
      <nav className="layout-nav" aria-label="Main navigation">
        <div className="nav-brand">Dojo CRM</div>
        <ul className="nav-links">
          {visibleItems.map((item) => (
            <li key={item.to}>
              <Link to={item.to}>{item.label}</Link>
            </li>
          ))}
        </ul>
        {user && (
          <div className="nav-user">
            <span>{user.name}</span>
            <button onClick={handleLogout}>Logout</button>
          </div>
        )}
      </nav>
      <main className="layout-content">
        <Outlet />
      </main>
    </div>
  )
}
