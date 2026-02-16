import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { listUsers } from '../api/users'
import type { Member } from '../api/users'

export function MemberListPage() {
  const { user } = useAuth()
  const [members, setMembers] = useState<Member[]>([])
  const [search, setSearch] = useState('')
  const [roleFilter, setRoleFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    listUsers()
      .then(setMembers)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load members'))
      .finally(() => setLoading(false))
  }, [])

  const filtered = members.filter((m) => {
    const matchesSearch =
      !search ||
      m.name.toLowerCase().includes(search.toLowerCase()) ||
      m.email.toLowerCase().includes(search.toLowerCase()) ||
      m.phone.includes(search)
    const matchesRole = !roleFilter || m.role === roleFilter
    const matchesStatus = !statusFilter || m.membership_status === statusFilter
    return matchesSearch && matchesRole && matchesStatus
  })

  if (loading) return <div>Loading members...</div>
  if (error) return <div role="alert">{error}</div>

  const isAdmin = user?.role === 'admin'

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2>Members</h2>
        {isAdmin && <Link to="/members/new">Add Member</Link>}
      </div>

      <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
        <input
          type="text"
          placeholder="Search by name, email, or phone"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          aria-label="Search members"
        />
        <select
          value={roleFilter}
          onChange={(e) => setRoleFilter(e.target.value)}
          aria-label="Filter by role"
        >
          <option value="">All roles</option>
          <option value="admin">Admin</option>
          <option value="instructor">Instructor</option>
          <option value="user">User</option>
        </select>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          aria-label="Filter by status"
        >
          <option value="">All statuses</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
          <option value="suspended">Suspended</option>
        </select>
      </div>

      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Email</th>
            <th>Role</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          {filtered.map((m) => (
            <tr key={m.id}>
              <td>
                <Link to={`/members/${m.id}`}>{m.name}</Link>
              </td>
              <td>{m.email}</td>
              <td>{m.role}</td>
              <td>{m.membership_status || '—'}</td>
            </tr>
          ))}
          {filtered.length === 0 && (
            <tr>
              <td colSpan={4}>No members found</td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  )
}
