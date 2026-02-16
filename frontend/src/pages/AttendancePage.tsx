import { useCallback, useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { listClasses, listClassTypes } from '../api/classes'
import type { Class, ClassType } from '../api/classes'
import { recordAttendance, getClassAttendance } from '../api/attendance'
import type { Attendance } from '../api/attendance'

export function AttendancePage() {
  const [classes, setClasses] = useState<Class[]>([])
  const [classTypes, setClassTypes] = useState<Map<number, ClassType>>(new Map())
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')

  const [selectedClassId, setSelectedClassId] = useState('')
  const [userId, setUserId] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const [attendance, setAttendance] = useState<Attendance[]>([])
  const [attendanceLoading, setAttendanceLoading] = useState(false)

  const fetchClasses = useCallback(async () => {
    try {
      const [classesData, typesData] = await Promise.all([
        listClasses(),
        listClassTypes(),
      ])
      setClasses(classesData)
      const typeMap = new Map<number, ClassType>()
      for (const ct of typesData) {
        typeMap.set(ct.id, ct)
      }
      setClassTypes(typeMap)
    } catch {
      setLoadError('Failed to load classes')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchClasses()
  }, [fetchClasses])

  const fetchAttendance = useCallback(async (classId: number) => {
    setAttendanceLoading(true)
    try {
      const data = await getClassAttendance(classId)
      setAttendance(data)
    } catch {
      setAttendance([])
    } finally {
      setAttendanceLoading(false)
    }
  }, [])

  async function handleClassSelect(classId: string) {
    setSelectedClassId(classId)
    setAttendance([])
    if (classId) {
      await fetchAttendance(Number(classId))
    }
  }

  async function handleCheckIn(e: FormEvent) {
    e.preventDefault()
    if (!selectedClassId || !userId) return

    setError('')
    setSuccess('')
    setSubmitting(true)
    try {
      await recordAttendance(Number(selectedClassId), { user_id: Number(userId) })
      setSuccess('Attendance recorded successfully')
      setUserId('')
      await fetchAttendance(Number(selectedClassId))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to record attendance')
    } finally {
      setSubmitting(false)
    }
  }

  function formatDateTime(iso: string): string {
    const d = new Date(iso)
    return d.toLocaleString(undefined, {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  if (loading) return <div>Loading...</div>

  return (
    <div>
      <h2>Attendance</h2>

      {loadError && <div role="alert">{loadError}</div>}

      <section>
        <h3>Record Attendance</h3>
        <form onSubmit={handleCheckIn}>
          <div>
            <label htmlFor="att-class">Class</label>
            <select
              id="att-class"
              value={selectedClassId}
              onChange={(e) => handleClassSelect(e.target.value)}
              required
            >
              <option value="">Select a class</option>
              {classes.map((cls) => (
                <option key={cls.id} value={cls.id}>
                  {classTypes.get(cls.class_type_id)?.name ?? 'Class'} — {formatDateTime(cls.start_time)}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label htmlFor="att-user-id">Member ID</label>
            <input
              id="att-user-id"
              type="number"
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              required
            />
          </div>
          {error && <div role="alert">{error}</div>}
          {success && <div role="status">{success}</div>}
          <button type="submit" disabled={submitting}>
            {submitting ? 'Recording...' : 'Check In'}
          </button>
        </form>
      </section>

      {selectedClassId && (
        <section>
          <h3>Class Attendance</h3>
          {attendanceLoading ? (
            <p>Loading attendance...</p>
          ) : attendance.length === 0 ? (
            <p>No attendance recorded for this class.</p>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>Member ID</th>
                  <th>Checked In</th>
                </tr>
              </thead>
              <tbody>
                {attendance.map((a) => (
                  <tr key={a.id}>
                    <td>{a.user_id}</td>
                    <td>{formatDateTime(a.checked_in_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </section>
      )}
    </div>
  )
}
