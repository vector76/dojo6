import { useCallback, useEffect, useState } from 'react'
import {
  createClassType,
  deleteClassType,
  listClassTypes,
  updateClassType,
} from '../api/classes'
import type { ClassType } from '../api/classes'

export default function ClassTypes() {
  const [classTypes, setClassTypes] = useState<ClassType[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [formName, setFormName] = useState('')
  const [formDescription, setFormDescription] = useState('')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)

  const fetchClassTypes = useCallback(async () => {
    try {
      setError(null)
      const data = await listClassTypes()
      setClassTypes(data)
    } catch {
      setError('Failed to load class types')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchClassTypes()
  }, [fetchClassTypes])

  function resetForm() {
    setFormName('')
    setFormDescription('')
    setEditingId(null)
  }

  function startEdit(ct: ClassType) {
    setEditingId(ct.id)
    setFormName(ct.name)
    setFormDescription(ct.description)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!formName.trim()) return

    setSaving(true)
    setError(null)
    try {
      if (editingId !== null) {
        await updateClassType(editingId, { name: formName.trim(), description: formDescription.trim() })
      } else {
        await createClassType({ name: formName.trim(), description: formDescription.trim() })
      }
      resetForm()
      await fetchClassTypes()
    } catch {
      setError(editingId !== null ? 'Failed to update class type' : 'Failed to create class type')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id: number) {
    setError(null)
    try {
      await deleteClassType(id)
      if (editingId === id) resetForm()
      await fetchClassTypes()
    } catch {
      setError('Failed to delete class type')
    }
  }

  if (loading) return <p>Loading class types...</p>

  return (
    <div>
      <h2>Class Types</h2>

      {error && <div role="alert">{error}</div>}

      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="ct-name">Name</label>
          <input
            id="ct-name"
            type="text"
            value={formName}
            onChange={(e) => setFormName(e.target.value)}
            required
          />
        </div>
        <div>
          <label htmlFor="ct-description">Description</label>
          <input
            id="ct-description"
            type="text"
            value={formDescription}
            onChange={(e) => setFormDescription(e.target.value)}
          />
        </div>
        <button type="submit" disabled={saving}>
          {editingId !== null ? 'Update' : 'Create'}
        </button>
        {editingId !== null && (
          <button type="button" onClick={resetForm}>Cancel</button>
        )}
      </form>

      {classTypes.length === 0 ? (
        <p>No class types yet.</p>
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Description</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {classTypes.map((ct) => (
                <tr key={ct.id}>
                  <td>{ct.name}</td>
                  <td>{ct.description}</td>
                  <td>
                    <button onClick={() => startEdit(ct)}>Edit</button>
                    <button onClick={() => handleDelete(ct.id)}>Delete</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
