import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { authAPI } from '../api'
import { Home } from 'lucide-react'

export default function RegisterPage() {
    const navigate = useNavigate()
    const [fullName, setFullName] = useState('')
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [confirm, setConfirm] = useState('')
    const [error, setError] = useState('')
    const [loading, setLoading] = useState(false)

    const inputStyle: React.CSSProperties = {
        width: '100%', padding: '8px 12px',
        background: 'rgba(255,255,255,0.02)',
        border: '1px solid rgba(255,255,255,0.08)',
        borderRadius: '6px',
        color: 'var(--text2)', fontSize: '15px',
        outline: 'none', transition: 'border-color 0.15s',
        boxSizing: 'border-box',
    }

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setError('')

        if (password !== confirm) {
            setError('Passwords do not match')
            return
        }
        if (password.length < 8) {
            setError('Password must be at least 8 characters')
            return
        }

        setLoading(true)
        try {
            await authAPI.register({ full_name: fullName, email, password })
            // After registration, redirect to login with success hint
            navigate('/login?registered=1')
        } catch (err: any) {
            const msg = err?.response?.data?.error || 'Registration failed. Please try again.'
            setError(msg)
        } finally {
            setLoading(false)
        }
    }

    const onFocus = (e: React.FocusEvent<HTMLInputElement>) =>
        (e.target.style.borderColor = 'rgba(113,112,255,0.5)')
    const onBlur = (e: React.FocusEvent<HTMLInputElement>) =>
        (e.target.style.borderColor = 'rgba(255,255,255,0.08)')

    return (
        <div style={{
            minHeight: '100vh', background: 'var(--bg)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            padding: '24px',
        }}>
            {/* Ambient glow */}
            <div style={{
                position: 'fixed', top: '20%', left: '50%', transform: 'translateX(-50%)',
                width: '500px', height: '300px',
                background: 'radial-gradient(ellipse, rgba(94,106,210,0.08) 0%, transparent 70%)',
                pointerEvents: 'none',
            }} />

            <div style={{ width: '100%', maxWidth: '380px', position: 'relative' }}>
                {/* Logo */}
                <div style={{ textAlign: 'center', marginBottom: '32px' }}>
                    <div style={{
                        display: 'inline-flex', alignItems: 'center', gap: '8px', marginBottom: '24px',
                    }}>
                        <div style={{
                            width: '28px', height: '28px', background: 'var(--accent)',
                            borderRadius: '7px', display: 'flex', alignItems: 'center',
                            justifyContent: 'center', color: '#fff',
                        }}>
                            <Home size={14} />
                        </div>
                        <span style={{ fontSize: '15px', fontWeight: 510, color: 'var(--text)', letterSpacing: '-0.165px' }}>
                            DormOS
                        </span>
                    </div>
                    <h1 style={{ fontSize: '24px', fontWeight: 510, color: 'var(--text)', letterSpacing: '-0.288px', marginBottom: '6px' }}>
                        Create your account
                    </h1>
                    <p style={{ fontSize: '14px', color: 'var(--text3)', letterSpacing: '-0.13px' }}>
                        Join your dormitory portal
                    </p>
                </div>

                {/* Form card */}
                <div style={{
                    background: 'rgba(255,255,255,0.02)',
                    border: '1px solid rgba(255,255,255,0.08)',
                    borderRadius: '12px', padding: '24px',
                }}>
                    <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>

                        {/* Full name */}
                        <div>
                            <label style={{ display: 'block', fontSize: '13px', fontWeight: 510, color: 'var(--text3)', marginBottom: '6px', letterSpacing: '-0.13px' }}>
                                Full Name
                            </label>
                            <input
                                type="text" value={fullName} onChange={e => setFullName(e.target.value)}
                                placeholder="John Doe" required style={inputStyle}
                                onFocus={onFocus} onBlur={onBlur}
                            />
                        </div>

                        {/* Email */}
                        <div>
                            <label style={{ display: 'block', fontSize: '13px', fontWeight: 510, color: 'var(--text3)', marginBottom: '6px', letterSpacing: '-0.13px' }}>
                                Email
                            </label>
                            <input
                                type="email" value={email} onChange={e => setEmail(e.target.value)}
                                placeholder="you@example.com" required style={inputStyle}
                                onFocus={onFocus} onBlur={onBlur}
                            />
                        </div>

                        {/* Password */}
                        <div>
                            <label style={{ display: 'block', fontSize: '13px', fontWeight: 510, color: 'var(--text3)', marginBottom: '6px', letterSpacing: '-0.13px' }}>
                                Password
                            </label>
                            <input
                                type="password" value={password} onChange={e => setPassword(e.target.value)}
                                placeholder="Min 8 characters" required style={inputStyle}
                                onFocus={onFocus} onBlur={onBlur}
                            />
                        </div>

                        {/* Confirm password */}
                        <div>
                            <label style={{ display: 'block', fontSize: '13px', fontWeight: 510, color: 'var(--text3)', marginBottom: '6px', letterSpacing: '-0.13px' }}>
                                Confirm Password
                            </label>
                            <input
                                type="password" value={confirm} onChange={e => setConfirm(e.target.value)}
                                placeholder="Repeat password" required style={inputStyle}
                                onFocus={onFocus} onBlur={onBlur}
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
                            {loading ? 'Creating account...' : 'Create account'}
                        </button>
                    </form>
                </div>

                {/* Sign in link */}
                <p style={{ textAlign: 'center', fontSize: '13px', color: 'var(--text3)', marginTop: '16px' }}>
                    Already have an account?{' '}
                    <Link to="/login" style={{
                        color: 'var(--accent-bright)', textDecoration: 'none',
                        fontWeight: 510,
                    }}
                        onMouseOver={e => ((e.target as HTMLElement).style.color = 'var(--accent-hover)')}
                        onMouseOut={e => ((e.target as HTMLElement).style.color = 'var(--accent-bright)')}
                    >
                        Sign in
                    </Link>
                </p>

                <p style={{ textAlign: 'center', fontSize: '12px', color: 'var(--text4)', marginTop: '8px' }}>
                    Dormitory Management System
                </p>
            </div>
        </div>
    )
}
