import { BrowserRouter, Routes, Route, useNavigate } from 'react-router-dom'
import { useCallback } from 'react'
import { AuthProvider } from './hooks/useAuth'
import { ProtectedRoute } from './components/ProtectedRoute'
import { Layout } from './components/Layout'
import { LoginPage } from './pages/LoginPage'
import { SetupPage } from './pages/SetupPage'
import { DashboardPage } from './pages/DashboardPage'
import { PaymentsPage } from './pages/PaymentsPage'
import { MyPaymentsPage } from './pages/MyPaymentsPage'
import ClassSchedule from './pages/ClassSchedule'
import ClassTypes from './pages/ClassTypes'
import { MemberListPage } from './pages/MemberListPage'
import { MemberDetailPage } from './pages/MemberDetailPage'
import { AddMemberPage } from './pages/AddMemberPage'

function AppRoutes() {
  const navigate = useNavigate()
  const onUnauthorized = useCallback(() => {
    navigate('/login')
  }, [navigate])

  return (
    <AuthProvider onUnauthorized={onUnauthorized}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/setup" element={<SetupPage />} />
        <Route
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route path="/" element={<DashboardPage />} />
          <Route path="/payments" element={<PaymentsPage />} />
          <Route path="/my-payments" element={<MyPaymentsPage />} />
          <Route path="/classes" element={<ClassSchedule />} />
          <Route path="/class-types" element={<ClassTypes />} />
          <Route path="/members" element={<MemberListPage />} />
          <Route path="/members/new" element={<AddMemberPage />} />
          <Route path="/members/:id" element={<MemberDetailPage />} />
        </Route>
      </Routes>
    </AuthProvider>
  )
}

function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  )
}

export default App
