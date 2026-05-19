import { createContext, useContext, useState, useEffect } from 'react'
import type { ReactNode } from 'react'
import { authAPI } from './api'

interface User {
    id: string
    name: string
    email: string
    room_number: string
    role: 'student' | 'manager' | 'admin'
}

interface AuthContextType {
    user: User | null
    loading: boolean
    login: (email: string, password: string) => Promise<void>
    logout: () => void
}

const AuthContext = createContext<AuthContextType>(null!)

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null)
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        const token = localStorage.getItem('token')
        if (token) {
            authAPI.me()
                .then(res => setUser({
                    id: res.data.user_id,
                    name: res.data.full_name,
                    email: res.data.email,
                    room_number: res.data.room_number || '',
                    role: res.data.role || 'student',
                }))
                .catch(() => localStorage.removeItem('token'))
                .finally(() => setLoading(false))
        } else {
            setLoading(false)
        }
    }, [])

    const login = async (email: string, password: string) => {
        const res = await authAPI.login(email, password)
        localStorage.setItem('token', res.data.access_token)
        // Fetch profile after login to populate user
        const profile = await authAPI.me()
        setUser({
            id: profile.data.user_id,
            name: profile.data.full_name,
            email: profile.data.email,
            room_number: profile.data.room_number || '',
            role: profile.data.role || 'student',
        })
    }

    const logout = () => {
        localStorage.removeItem('token')
        setUser(null)
    }

    return (
        <AuthContext.Provider value={{ user, loading, login, logout }}>
    {children}
    </AuthContext.Provider>
)
}

export const useAuth = () => useContext(AuthContext)
