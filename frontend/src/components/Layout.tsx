import { NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'

const S = {
    sidebar: {
        width: '220px', flexShrink: 0,
        background: 'var(--panel)',
        borderRight: '1px solid var(--border-subtle)',
        display: 'flex', flexDirection: 'column' as const,
        padding: '12px 8px',
        position: 'fixed' as const, height: '100vh', top: 0, left: 0, zIndex: 10,
    },
    logo: {
        display: 'flex', alignItems: 'center', gap: '8px',
        padding: '8px 10px', marginBottom: '16px',
    },
    logoIcon: {
        width: '22px', height: '22px',
        background: 'var(--accent)',
        borderRadius: '6px',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        fontSize: '12px',
    },
    logoText: {
        fontSize: '14px', fontWeight: 510, color: 'var(--text)',
        letterSpacing: '-0.182px',
    },
    section: { marginBottom: '4px' },
    sectionLabel: {
        fontSize: '11px', fontWeight: 510, color: 'var(--text4)',
        padding: '4px 10px', letterSpacing: '0.05em', textTransform: 'uppercase' as const,
        marginBottom: '2px',
    },
    divider: { height: '1px', background: 'var(--border-subtle)', margin: '8px 4px' },
    user: { borderTop: '1px solid var(--border-subtle)', paddingTop: '8px', marginTop: 'auto' },
    userInfo: {
        display: 'flex', alignItems: 'center', gap: '8px',
        padding: '8px 10px', borderRadius: '6px',
    },
    avatar: {
        width: '24px', height: '24px', borderRadius: '50%',
        background: 'var(--accent)', display: 'flex',
        alignItems: 'center', justifyContent: 'center',
        fontSize: '11px', fontWeight: 590, color: '#fff', flexShrink: 0,
    },
    logoutBtn: {
        width: '100%', padding: '6px 10px',
        background: 'transparent',
        border: 'none',
        borderRadius: '6px',
        color: 'var(--text4)', fontSize: '13px',
        cursor: 'pointer', textAlign: 'left' as const,
        transition: 'color 0.15s, background 0.15s',
    },
}

const navItems = [
    { to: '/dashboard', icon: '⚡', label: 'Dashboard' },
    { to: '/issues', icon: '🔧', label: 'Issues' },
    { to: '/documents', icon: '📄', label: 'Documents' },
    { to: '/activities', icon: '🏆', label: 'Activities' },
]

export default function Layout({ children }: { children: React.ReactNode }) {
    const { user, logout } = useAuth()
    const navigate = useNavigate()

    const navLinkStyle = (isActive: boolean) => ({
        display: 'flex', alignItems: 'center', gap: '8px',
        padding: '6px 10px', borderRadius: '6px',
        textDecoration: 'none',
        fontSize: '13px', fontWeight: isActive ? 510 : 400,
        color: isActive ? 'var(--text)' : 'var(--text3)',
        background: isActive ? 'rgba(255,255,255,0.05)' : 'transparent',
        transition: 'all 0.1s',
        letterSpacing: '-0.13px',
    })

    return (
        <div style={{ display: 'flex', minHeight: '100vh' }}>
            <aside style={S.sidebar}>
                {/* Logo */}
                <div style={S.logo}>
                    <div style={S.logoIcon}>🏠</div>
                    <span style={S.logoText}>DormOS</span>
                </div>

                {/* Nav */}
                <nav style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '1px' }}>
                    {navItems.map(item => (
                        <NavLink key={item.to} to={item.to} style={({ isActive }) => navLinkStyle(isActive)}>
                            <span style={{ fontSize: '13px', opacity: 0.8 }}>{item.icon}</span>
                            {item.label}
                        </NavLink>
                    ))}

                    {(user?.role === 'admin' || user?.role === 'manager') && (
                        <>
                            <div style={S.divider} />
                            <NavLink to="/admin" style={({ isActive }) => navLinkStyle(isActive)}>
                                <span style={{ fontSize: '13px', opacity: 0.8 }}>⚙️</span>
                                Admin
                            </NavLink>
                        </>
                    )}
                </nav>

                {/* User */}
                <div style={S.user}>
                    <div style={S.userInfo}>
                        <div style={S.avatar}>{user?.name?.[0]?.toUpperCase() || 'U'}</div>
                        <div style={{ flex: 1, minWidth: 0 }}>
                            <div style={{ fontSize: '13px', fontWeight: 510, color: 'var(--text)', letterSpacing: '-0.13px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                                {user?.name || 'User'}
                            </div>
                            <div style={{ fontSize: '11px', color: 'var(--text4)' }}>Room {user?.room_number || '—'}</div>
                        </div>
                    </div>
                    <button
                        onClick={() => { logout(); navigate('/login') }}
                        style={S.logoutBtn}
                        onMouseOver={e => { (e.currentTarget.style.color = 'var(--text2)'); (e.currentTarget.style.background = 'rgba(255,255,255,0.03)') }}
                        onMouseOut={e => { (e.currentTarget.style.color = 'var(--text4)'); (e.currentTarget.style.background = 'transparent') }}
                    >
                        Sign out
                    </button>
                </div>
            </aside>

            <main style={{ marginLeft: '220px', flex: 1, padding: '32px 40px', minHeight: '100vh' }}>
                {children}
            </main>
        </div>
    )
}
