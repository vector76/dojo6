import { useState } from 'react'
import type { FormEvent } from 'react'
import { recordPayment, getPaymentHistory, getBalance } from '../api/payments'
import type { Payment, BalanceInfo } from '../api/payments'

export function PaymentsPage() {
  const [userId, setUserId] = useState('')
  const [amount, setAmount] = useState('')
  const [date, setDate] = useState(new Date().toISOString().split('T')[0])
  const [note, setNote] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const [historyUserId, setHistoryUserId] = useState('')
  const [payments, setPayments] = useState<Payment[]>([])
  const [balanceInfo, setBalanceInfo] = useState<BalanceInfo | null>(null)
  const [historyLoading, setHistoryLoading] = useState(false)
  const [historyError, setHistoryError] = useState('')

  const handleRecordPayment = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setSuccess('')
    setSubmitting(true)
    try {
      await recordPayment({
        user_id: Number(userId),
        amount: Number(amount),
        date,
        note,
      })
      setSuccess('Payment recorded successfully')
      setAmount('')
      setNote('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to record payment')
    } finally {
      setSubmitting(false)
    }
  }

  const handleViewHistory = async (e: FormEvent) => {
    e.preventDefault()
    setHistoryError('')
    setHistoryLoading(true)
    try {
      const uid = Number(historyUserId)
      const [paymentList, balance] = await Promise.all([
        getPaymentHistory(uid),
        getBalance(uid),
      ])
      setPayments(paymentList)
      setBalanceInfo(balance)
    } catch (err) {
      setHistoryError(err instanceof Error ? err.message : 'Failed to load history')
      setPayments([])
      setBalanceInfo(null)
    } finally {
      setHistoryLoading(false)
    }
  }

  return (
    <div>
      <h2>Payments</h2>

      <section>
        <h3>Record Payment</h3>
        <form onSubmit={handleRecordPayment}>
          <div>
            <label htmlFor="pay-user-id">Member ID</label>
            <input
              id="pay-user-id"
              type="number"
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              required
            />
          </div>
          <div>
            <label htmlFor="pay-amount">Amount</label>
            <input
              id="pay-amount"
              type="number"
              step="0.01"
              min="0"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              required
            />
          </div>
          <div>
            <label htmlFor="pay-date">Date</label>
            <input
              id="pay-date"
              type="date"
              value={date}
              onChange={(e) => setDate(e.target.value)}
              required
            />
          </div>
          <div>
            <label htmlFor="pay-note">Note</label>
            <input
              id="pay-note"
              type="text"
              value={note}
              onChange={(e) => setNote(e.target.value)}
            />
          </div>
          {error && <div role="alert">{error}</div>}
          {success && <div role="status">{success}</div>}
          <button type="submit" disabled={submitting}>
            {submitting ? 'Recording...' : 'Record Payment'}
          </button>
        </form>
      </section>

      <section>
        <h3>Payment History</h3>
        <form onSubmit={handleViewHistory}>
          <div>
            <label htmlFor="hist-user-id">Member ID</label>
            <input
              id="hist-user-id"
              type="number"
              value={historyUserId}
              onChange={(e) => setHistoryUserId(e.target.value)}
              required
            />
          </div>
          <button type="submit" disabled={historyLoading}>
            {historyLoading ? 'Loading...' : 'View History'}
          </button>
        </form>
        {historyError && <div role="alert">{historyError}</div>}

        {balanceInfo && (
          <div data-testid="balance-summary">
            <h4>Balance Summary</h4>
            <p>Expected: ${balanceInfo.expected_balance.toFixed(2)}</p>
            <p>Total Paid: ${balanceInfo.total_payments.toFixed(2)}</p>
            <p>Outstanding: ${balanceInfo.outstanding_balance.toFixed(2)}</p>
          </div>
        )}

        {payments.length > 0 && (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Amount</th>
                  <th>Note</th>
                </tr>
              </thead>
              <tbody>
                {payments.map((p) => (
                  <tr key={p.id}>
                    <td>{p.date}</td>
                    <td>${p.amount.toFixed(2)}</td>
                    <td>{p.note}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  )
}
