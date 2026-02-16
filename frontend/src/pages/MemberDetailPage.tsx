import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { getUser, updateUser, deleteUser, resetPassword, setExpectedBalance } from '../api/users'
import type { Member, UpdateMemberRequest } from '../api/users'

export function MemberDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { user: currentUser } = useAuth()
  const navigate = useNavigate()

  const [member, setMember] = useState<Member | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const [saveMessage, setSaveMessage] = useState('')

  // Edit form state
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [role, setRole] = useState<'admin' | 'instructor' | 'user'>('user')
  const [membershipType, setMembershipType] = useState('')
  const [membershipStatus, setMembershipStatus] = useState('')
  const [emergencyContact, setEmergencyContact] = useState('')
  const [joinDate, setJoinDate] = useState('')

  // Password reset state
  const [newPassword, setNewPassword] = useState('')
  const [passwordMessage, setPasswordMessage] = useState('')

  // Balance state
  const [balanceInput, setBalanceInput] = useState('')
  const [balanceMessage, setBalanceMessage] = useState('')

  const isAdmin = currentUser?.role === 'admin'
  const isSelf = currentUser?.id === id
  const canEdit = isAdmin || isSelf

  useEffect(() => {
    if (!id) return
    getUser(id)
      .then((m) => {
        setMember(m)
        setName(m.name)
        setEmail(m.email)
        setPhone(m.phone)
        setRole(m.role)
        setMembershipType(m.membership_type)
        setMembershipStatus(m.membership_status)
        setEmergencyContact(m.emergency_contact)
        setJoinDate(m.join_date)
        setBalanceInput(String(m.expected_balance))
      })
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load member'))
      .finally(() => setLoading(false))
  }, [id])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!id) return
    setSaving(true)
    setSaveMessage('')
    try {
      const data: UpdateMemberRequest = {
        name,
        email,
        phone,
        role,
        membership_type: membershipType,
        membership_status: membershipStatus,
        emergency_contact: emergencyContact,
        join_date: joinDate,
      }
      const updated = await updateUser(id, data)
      setMember(updated)
      setSaveMessage('Saved successfully')
    } catch (err) {
      setSaveMessage(err instanceof Error ? err.message : 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!id || !confirm('Are you sure you want to delete this member?')) return
    try {
      await deleteUser(id)
      navigate('/members')
    } catch (err) {
      setSaveMessage(err instanceof Error ? err.message : 'Failed to delete')
    }
  }

  const handlePasswordReset = async (e: FormEvent) => {
    e.preventDefault()
    if (!id) return
    setPasswordMessage('')
    try {
      await resetPassword(id, newPassword)
      setNewPassword('')
      setPasswordMessage('Password updated')
    } catch (err) {
      setPasswordMessage(err instanceof Error ? err.message : 'Failed to reset password')
    }
  }

  const handleSetBalance = async (e: FormEvent) => {
    e.preventDefault()
    if (!id) return
    setBalanceMessage('')
    try {
      const updated = await setExpectedBalance(id, parseFloat(balanceInput))
      setMember(updated)
      setBalanceMessage('Balance updated')
    } catch (err) {
      setBalanceMessage(err instanceof Error ? err.message : 'Failed to update balance')
    }
  }

  if (loading) return <div>Loading member...</div>
  if (error) return <div role="alert">{error}</div>
  if (!member) return <div>Member not found</div>

  return (
    <div>
      <h2>Member Detail</h2>

      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="name">Name</label>
          <input
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            disabled={!canEdit}
          />
        </div>
        <div>
          <label htmlFor="email">Email</label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            disabled={!canEdit}
          />
        </div>
        <div>
          <label htmlFor="phone">Phone</label>
          <input
            id="phone"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            required
            disabled={!canEdit}
          />
        </div>
        {isAdmin && (
          <div>
            <label htmlFor="role">Role</label>
            <select id="role" value={role} onChange={(e) => setRole(e.target.value as 'admin' | 'instructor' | 'user')}>
              <option value="user">User</option>
              <option value="instructor">Instructor</option>
              <option value="admin">Admin</option>
            </select>
          </div>
        )}
        <div>
          <label htmlFor="membership_type">Membership Type</label>
          <select id="membership_type" value={membershipType} onChange={(e) => setMembershipType(e.target.value)} disabled={!isAdmin}>
            <option value="">None</option>
            <option value="monthly">Monthly</option>
            <option value="annual">Annual</option>
            <option value="drop-in">Drop-in</option>
          </select>
        </div>
        <div>
          <label htmlFor="membership_status">Membership Status</label>
          <select id="membership_status" value={membershipStatus} onChange={(e) => setMembershipStatus(e.target.value)} disabled={!isAdmin}>
            <option value="">None</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="suspended">Suspended</option>
          </select>
        </div>
        <div>
          <label htmlFor="emergency_contact">Emergency Contact</label>
          <input
            id="emergency_contact"
            value={emergencyContact}
            onChange={(e) => setEmergencyContact(e.target.value)}
            disabled={!canEdit}
          />
        </div>
        <div>
          <label htmlFor="join_date">Join Date</label>
          <input
            id="join_date"
            type="date"
            value={joinDate}
            onChange={(e) => setJoinDate(e.target.value)}
            disabled={!isAdmin}
          />
        </div>
        {canEdit && (
          <button type="submit" disabled={saving}>
            {saving ? 'Saving...' : 'Save Changes'}
          </button>
        )}
        {saveMessage && <div role="status">{saveMessage}</div>}
      </form>

      {isAdmin && (
        <>
          <hr />
          <h3>Set Expected Balance</h3>
          <form onSubmit={handleSetBalance}>
            <div>
              <label htmlFor="expected_balance">Expected Balance</label>
              <input
                id="expected_balance"
                type="number"
                step="0.01"
                value={balanceInput}
                onChange={(e) => setBalanceInput(e.target.value)}
              />
            </div>
            <button type="submit">Update Balance</button>
            {balanceMessage && <div role="status">{balanceMessage}</div>}
          </form>

          <hr />
          <h3>Reset Password</h3>
          <form onSubmit={handlePasswordReset}>
            <div>
              <label htmlFor="new_password">New Password</label>
              <input
                id="new_password"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
              />
            </div>
            <button type="submit">Reset Password</button>
            {passwordMessage && <div role="status">{passwordMessage}</div>}
          </form>

          <hr />
          <button onClick={handleDelete} style={{ color: 'red' }}>
            Delete Member
          </button>
        </>
      )}
    </div>
  )
}
