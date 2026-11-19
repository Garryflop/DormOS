import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './AuthContext.tsx'
import Layout from './components/Layout'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import DashboardPage from './pages/DashboardPage.tsx'
import IssuesPage from './pages/IssuesPage'
import DocumentsPage from './pages/DocumentsPage'
import ActivitiesPage from './pages/ActivitiesPage'
import AdminPage from './pages/AdminPage'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return (
      <div style={{
        minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center',
        background: 'var(--bg)', color: 'var(--text3)', fontSize: '14px',
        fontFeatureSettings: "'cv01','ss03'",
      }}>
        Loading...
      </div>
  )
  if (!user) return <Navigate to="/login" replace />
  return <Layout>{children}</Layout>
}

function AdminRoute({ children }: { children: React.ReactNode }) {
  const { user } = useAuth()
  const adminRoles = ['admin', 'manager']
  if (!user || !adminRoles.includes(user.role)) return <Navigate to="/dashboard" replace />
  return <>{children}</>
}

export default function App() {
  return (
      <BrowserRouter>
        <AuthProvider>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<PrivateRoute><DashboardPage /></PrivateRoute>} />
            <Route path="/issues" element={<PrivateRoute><IssuesPage /></PrivateRoute>} />
            <Route path="/documents" element={<PrivateRoute><DocumentsPage /></PrivateRoute>} />
            <Route path="/activities" element={<PrivateRoute><ActivitiesPage /></PrivateRoute>} />
            <Route path="/admin" element={<PrivateRoute><AdminRoute><AdminPage /></AdminRoute></PrivateRoute>} />
          </Routes>
        </AuthProvider>
      </BrowserRouter>
  )
}
