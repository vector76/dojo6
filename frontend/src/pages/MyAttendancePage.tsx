import { useEffect, useState } from 'react'
import { useAuth } from '../hooks/useAuth'
import { getUserAttendance } from '../api/attendance'
import type { Attendance } from '../api/attendance'

export function MyAttendancePage() {
  const { user } = useAuth()
  const [attendance, setAttendance] = useState<Attendance[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!user) return

    let cancelled = false
    getUserAttendance(Number(user.id))
      .then((data) => {
        if (!cancelled) {
          setAttendance(data)
          setLoading(false)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load attendance')
          setLoading(false)
        }
      })
    return () => {
      cancelled = true
    }
  }, [user])

  if (loading) {
    return <div>Loading...</div>
  }

  function formatDateTime(iso: string): string {
    const d = new Date(iso)
    return d.toLocaleString(undefined, {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <div>
      <h2>My Attendance</h2>

      {error && <div role="alert">{error}</div>}

      {attendance.length === 0 ? (
        <p>No attendance history found.</p>
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Class ID</th>
                <th>Checked In</th>
              </tr>
            </thead>
            <tbody>
              {attendance.map((a) => (
                <tr key={a.id}>
                  <td>{a.class_id}</td>
                  <td>{formatDateTime(a.checked_in_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
