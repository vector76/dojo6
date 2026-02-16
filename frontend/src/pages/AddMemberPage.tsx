import { useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { createUser } from '../api/users'
import type { CreateMemberRequest } from '../api/users'

export function AddMemberPage() {
  const navigate = useNavigate()
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<'admin' | 'instructor' | 'user'>('user')
  const [membershipType, setMembershipType] = useState('')
  const [membershipStatus, setMembershipStatus] = useState('')
  const [emergencyContact, setEmergencyContact] = useState('')
  const [joinDate, setJoinDate] = useState('')

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      const data: CreateMemberRequest = {
        name,
        email,
        phone,
        password,
        role,
        membership_type: membershipType,
        membership_status: membershipStatus,
        emergency_contact: emergencyContact,
        join_date: joinDate,
      }
      const created = await createUser(data)
      navigate(`/members/${created.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create member')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div>
      <h2>Add Member</h2>
      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="name">Name</label>
          <input id="name" value={name} onChange={(e) => setName(e.target.value)} required />
        </div>
        <div>
          <label htmlFor="email">Email</label>
          <input id="email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </div>
        <div>
          <label htmlFor="phone">Phone</label>
          <input id="phone" value={phone} onChange={(e) => setPhone(e.target.value)} required />
        </div>
        <div>
          <label htmlFor="password">Password</label>
          <input id="password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
        </div>
        <div>
          <label htmlFor="role">Role</label>
          <select id="role" value={role} onChange={(e) => setRole(e.target.value as 'admin' | 'instructor' | 'user')}>
            <option value="user">User</option>
            <option value="instructor">Instructor</option>
            <option value="admin">Admin</option>
          </select>
        </div>
        <div>
          <label htmlFor="membership_type">Membership Type</label>
          <select id="membership_type" value={membershipType} onChange={(e) => setMembershipType(e.target.value)}>
            <option value="">None</option>
            <option value="monthly">Monthly</option>
            <option value="annual">Annual</option>
            <option value="drop-in">Drop-in</option>
          </select>
        </div>
        <div>
          <label htmlFor="membership_status">Membership Status</label>
          <select id="membership_status" value={membershipStatus} onChange={(e) => setMembershipStatus(e.target.value)}>
            <option value="">None</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="suspended">Suspended</option>
          </select>
        </div>
        <div>
          <label htmlFor="emergency_contact">Emergency Contact</label>
          <input id="emergency_contact" value={emergencyContact} onChange={(e) => setEmergencyContact(e.target.value)} />
        </div>
        <div>
          <label htmlFor="join_date">Join Date</label>
          <input id="join_date" type="date" value={joinDate} onChange={(e) => setJoinDate(e.target.value)} />
        </div>
        {error && <div role="alert">{error}</div>}
        <button type="submit" disabled={submitting}>
          {submitting ? 'Creating...' : 'Create Member'}
        </button>
      </form>
    </div>
  )
}
