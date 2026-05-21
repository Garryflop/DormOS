import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { authAPI } from '../api'

export default function RegisterPage() {
    const navigate = useNavigate()
    const [form, setForm] = useState({
        full_name: '',
        email: '',
        password: '',
        room_number: '',
    })
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
            await authAPI.register(form)
            navigate('/login')
        } catch (err: any) {
            setError(err?.response?.data?.error || 'Registration failed')
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
            <div style={{
                position: 'fixed', top: '20%', left: '50%', transform: 'translateX(-50%)',
                width: '500px', height: '300px',
                background: 'radial-gradient(ellipse, rgba(94,106,210,0.08) 0%, transparent 70%)',
                pointerEvents: 'none',
            }} />

            <div style={{ width: '100%', maxWidth: '360px', position: 'relative' }}>
                {/* Logo */}
                <div style={{ textAlign: 'center', marginBottom: '32px' }}>
                    <div style={{ display: 'inline-flex', alignItems: 'center', gap: '8px', marginBottom: '24px' }}>
                        <div style={{
                            width: '28px', height: '28px', background: 'var(--accent)',
                            borderRadius: '7px', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '14px',
                        }}>🏠</div>
                        <span style={{ fontSize: '15px', fontWeight: 510, color: 'var(--text)', letterSpacing: '-0.165px' }}>DormOS</span>
                    </div>
                    <h1 style={{ fontSize: '24px', fontWeight: 510, color: 'var(--text)', letterSpacing: '-0.288px', marginBottom: '6px' }}>
                        Create account
                    </h1>
                    <p style={{ fontSize: '14px', color: 'var(--text3)', letterSpacing: '-0.13px' }}>
                        Join your dormitory portal
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
                            <label style={{ display: 'block', fontSize: '13px', fontWeight: 510, color: 'var(--text3)', marginBottom: '6px' }}>
                                Full name
                            </label>
                            <input
                                type="text" value={form.full_name} required
                                onChange={e => setForm({ ...form, full_name: e.target.value })}
                                placeholder="John Doe" style={inputStyle}
                                onFocus={e => e.target.style.borderColor = 'rgba(113,112,255,0.5)'}
                                onBlur={e => e.target.style.borderColor = 'rgba(255,255,255,0.08)'}
                            />
                        </div>

                        <div>
                            <label style={{ display: 'block', fontSize: '13px', fontWeight: 510, color: 'var(--text3)', marginBottom: '6px' }}>
                                Email
                            </label>
                            <input
                                type="email" value={form.email} required
                                onChange={e => setForm({ ...form, email: e.target.value })}
                                placeholder="you@example.com" style={inputStyle}
                                onFocus={e => e.target.style.borderColor = 'rgba(113,112,255,0.5)'}
                                onBlur={e => e.target.style.borderColor = 'rgba(255,255,255,0.08)'}
                            />
                        </div>

                        <div>
                            <label style={{ display: 'block', fontSize: '13px', fontWeight: 510, color: 'var(--text3)', marginBottom: '6px' }}>
                                Password
                            </label>
                            <input
                                type="password" value={form.password} required
                                onChange={e => setForm({ ...form, password: e.target.value })}
                                placeholder="••••••••" style={inputStyle}
                                onFocus={e => e.target.style.borderColor = 'rgba(113,112,255,0.5)'}
                                onBlur={e => e.target.style.borderColor = 'rgba(255,255,255,0.08)'}
                            />
                        </div>

                        <div>
                            <label style={{ display: 'block', fontSize: '13px', fontWeight: 510, color: 'var(--text3)', marginBottom: '6px' }}>
                                Room number
                            </label>
                            <input
                                type="text" value={form.room_number} required
                                onChange={e => setForm({ ...form, room_number: e.target.value })}
                                placeholder="305" style={inputStyle}
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
                            }}
                            onMouseOver={e => { if (!loading) (e.currentTarget.style.background = 'var(--accent-hover)') }}
                            onMouseOut={e => { if (!loading) (e.currentTarget.style.background = 'var(--accent)') }}
                        >
                            {loading ? 'Creating account...' : 'Create account'}
                        </button>
                    </form>
                </div>

                <p style={{ textAlign: 'center', fontSize: '13px', color: 'var(--text4)', marginTop: '16px' }}>
                    Already have an account?{' '}
                    <Link to="/login" style={{ color: 'var(--accent-bright)', textDecoration: 'none' }}>
                        Sign in
                    </Link>
                </p>
            </div>
        </div>
    )
}
