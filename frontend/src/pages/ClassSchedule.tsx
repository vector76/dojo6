import { useCallback, useEffect, useState } from 'react'
import { Class, ClassType, listClasses, listClassTypes } from '../api/classes'

function formatDate(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleDateString(undefined, {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

function formatTime(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function sortClassesByDate(classes: Class[]): Class[] {
  return [...classes].sort(
    (a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime(),
  )
}

export default function ClassSchedule() {
  const [classes, setClasses] = useState<Class[]>([])
  const [classTypes, setClassTypes] = useState<Map<number, ClassType>>(new Map())
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchData = useCallback(async () => {
    try {
      setError(null)
      const [classesData, typesData] = await Promise.all([
        listClasses(),
        listClassTypes(),
      ])
      setClasses(sortClassesByDate(classesData))
      const typeMap = new Map<number, ClassType>()
      for (const ct of typesData) {
        typeMap.set(ct.id, ct)
      }
      setClassTypes(typeMap)
    } catch {
      setError('Failed to load schedule')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  if (loading) return <p>Loading schedule...</p>

  return (
    <div>
      <h2>Class Schedule</h2>

      {error && <p role="alert" style={{ color: 'red' }}>{error}</p>}

      {classes.length === 0 ? (
        <p>No classes scheduled.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Date</th>
              <th>Time</th>
              <th>Type</th>
              <th>Duration</th>
              <th>Capacity</th>
            </tr>
          </thead>
          <tbody>
            {classes.map((cls) => (
              <tr key={cls.id}>
                <td>{formatDate(cls.start_time)}</td>
                <td>{formatTime(cls.start_time)}</td>
                <td>{classTypes.get(cls.class_type_id)?.name ?? 'Unknown'}</td>
                <td>{cls.duration_minutes} min</td>
                <td>{cls.capacity}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
