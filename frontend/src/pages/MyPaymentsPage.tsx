import { useEffect, useState } from 'react'
import { useAuth } from '../hooks/useAuth'
import { getPaymentHistory, getBalance } from '../api/payments'
import type { Payment, BalanceInfo } from '../api/payments'

export function MyPaymentsPage() {
  const { user } = useAuth()
  const [payments, setPayments] = useState<Payment[]>([])
  const [balanceInfo, setBalanceInfo] = useState<BalanceInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!user) return

    let cancelled = false
    Promise.all([getPaymentHistory(Number(user.id)), getBalance(Number(user.id))])
      .then(([paymentList, balance]) => {
        if (!cancelled) {
          setPayments(paymentList)
          setBalanceInfo(balance)
          setLoading(false)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load payments')
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

  return (
    <div>
      <h2>My Payments</h2>

      {error && <div role="alert">{error}</div>}

      {balanceInfo && (
        <div data-testid="balance-summary">
          <h3>Balance Summary</h3>
          <p>Expected: ${balanceInfo.expected_balance.toFixed(2)}</p>
          <p>Total Paid: ${balanceInfo.total_payments.toFixed(2)}</p>
          <p>Outstanding: ${balanceInfo.outstanding_balance.toFixed(2)}</p>
        </div>
      )}

      {payments.length === 0 ? (
        <p>No payment history found.</p>
      ) : (
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
      )}
    </div>
  )
}
