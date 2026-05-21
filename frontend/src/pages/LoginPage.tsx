import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'

export default function LoginPage() {
    const { login } = useAuth()
    const navigate = useNavigate()
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [error, setError] = useState('')
    const [loading, setLoading] = useState(false)

    const inputStyle = {
        width: '100%', padding: '8px 12px',
        background: 'rgba(255,255,255,0.02)',
        border: '1px solid rgba(255,255,255,0.08)',
        borderRadius: '6px',
        color: 'var(--text2)', fontSize: '15px',
        outline: 'none', transition: 'border-color 0.15s',
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setError('')
        setLoading(true)
        try {
            await login(email, password)
            navigate('/dashboard')
        } catch {
            setError('Invalid email or password')
        } finally {
            setLoading(false)
        }
    }

    return (
        <div style={{
            minHeight: '100vh', background: 'var(--bg)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            padding: '24px',
        }}>
            {/* Glow */}
            <div style={{
                position: 'fixed', top: '20%', left: '50%', transform: 'translateX(-50%)',
                width: '500px', height: '300px',
                background: 'radial-gradient(ellipse, rgba(94,106,210,0.08) 0%, transparent 70%)',
                pointerEvents: 'none',
            }} />

            <div style={{ width: '100%', maxWidth: '360px', position: 'relative' }}>
                {/* Logo */}
                <div style={{ textAlign: 'center', marginBottom: '32px' }}>
                    <div style={{
                        display: 'inline-flex', alignItems: 'center', gap: '8px', marginBottom: '24px',
                    }}>
                        <div style={{
                            width: '28px', height: '28px', background: 'var(--accent)',
                            borderRadius: '7px', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '14px',
                        }}>🏠</div>
                        <span style={{ fontSize: '15px', fontWeight: 510, color: 'var(--text)', letterSpacing: '-0.165px' }}>DormOS</span>
                    </div>
                    <h1 style={{ fontSize: '24px', fontWeight: 510, color: 'var(--text)', letterSpacing: '-0.288px', marginBottom: '6px' }}>
                        Sign in
                    </h1>
                    <p style={{ fontSize: '14px', color: 'var(--text3)', letterSpacing: '-0.13px' }}>
                        Access your dormitory portal
                    </p>
                </div>

                {/* Form */}
                <div style={{
                    background: 'rgba(255,255,255,0.02)',
                    border: '1px solid rgba(255,255,255,0.08)',
                    borderRadius: '12px', padding: '24px',
                }}>
                    <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                        <div>
                            <label style={{ display: 'block', fontSize: '13px', fontWeight: 510, color: 'var(--text3)', marginBottom: '6px', letterSpacing: '-0.13px' }}>
                                Email
                            </label>
                            <input
                                type="email" value={email} onChange={e => setEmail(e.target.value)}
                                placeholder="you@example.com" required style={inputStyle}
                                onFocus={e => e.target.style.borderColor = 'rgba(113,112,255,0.5)'}
                                onBlur={e => e.target.style.borderColor = 'rgba(255,255,255,0.08)'}
                            />
                        </div>

                        <div>
                            <label style={{ display: 'block', fontSize: '13px', fontWeight: 510, color: 'var(--text3)', marginBottom: '6px', letterSpacing: '-0.13px' }}>
                                Password
                            </label>
                            <input
                                type="password" value={password} onChange={e => setPassword(e.target.value)}
                                placeholder="••••••••" required style={inputStyle}
                                onFocus={e => e.target.style.borderColor = 'rgba(113,112,255,0.5)'}
                                onBlur={e => e.target.style.borderColor = 'rgba(255,255,255,0.08)'}
                            />
                        </div>

                        {error && (
                            <div style={{
                                padding: '8px 12px',
                                background: 'rgba(255,59,48,0.08)',
                                border: '1px solid rgba(255,59,48,0.2)',
                                borderRadius: '6px',
                                color: '#ff6b6b', fontSize: '13px',
                            }}>
                                {error}
                            </div>
                        )}

                        <button
                            type="submit" disabled={loading}
                            style={{
                                width: '100%', padding: '8px 16px',
                                background: loading ? 'rgba(94,106,210,0.5)' : 'var(--accent)',
                                border: 'none', borderRadius: '6px',
                                color: '#fff', fontSize: '14px', fontWeight: 510,
                                cursor: loading ? 'not-allowed' : 'pointer',
                                transition: 'background 0.15s', marginTop: '4px',
                                letterSpacing: '-0.13px',
                            }}
                            onMouseOver={e => { if (!loading) (e.currentTarget.style.background = 'var(--accent-hover)') }}
                            onMouseOut={e => { if (!loading) (e.currentTarget.style.background = 'var(--accent)') }}
                        >
                            {loading ? 'Signing in...' : 'Continue'}
                        </button>
                    </form>
                </div>

                <p style={{ textAlign: 'center', fontSize: '12px', color: 'var(--text4)', marginTop: '16px' }}>
                    Dormitory Management System
                </p>
            </div>
        </div>
    )
}
