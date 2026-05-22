import { useEffect, useState } from 'react'
import { NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import { notificationsAPI } from '../api'
import { 
	LayoutDashboard, 
	Wrench, 
	FileText, 
	Trophy, 
	Settings, 
	Home, 
	LogOut,
	Bell,
	CheckCheck
} from 'lucide-react'

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
		width: '24px', height: '24px',
		color: 'var(--accent)',
		display: 'flex', alignItems: 'center', justifyContent: 'center',
	},
	logoText: {
		fontSize: '15px', fontWeight: 600, color: 'var(--text)',
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
		cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '8px',
		textAlign: 'left' as const,
		transition: 'color 0.15s, background 0.15s',
	},
}

const navItems = [
	{ to: '/dashboard', icon: LayoutDashboard, label: 'Dashboard' },
	{ to: '/issues', icon: Wrench, label: 'Issues' },
	{ to: '/documents', icon: FileText, label: 'Documents' },
	{ to: '/activities', icon: Trophy, label: 'Activities' },
]

export default function Layout({ children }: { children: React.ReactNode }) {
	const { user, logout } = useAuth()
	const navigate = useNavigate()
	const [notifications, setNotifications] = useState<any[]>([])
	const [showNotifications, setShowNotifications] = useState(false)

	const loadNotifications = async () => {
		if (!user?.id) return
		try {
			const res = await notificationsAPI.list(user.id)
			setNotifications(res.data.notifications || [])
		} catch (err) {
			console.error('Failed to load notifications:', err)
		}
	}

	useEffect(() => {
		if (user?.id) {
			loadNotifications()
			const interval = setInterval(loadNotifications, 10000)
			return () => clearInterval(interval)
		}
	}, [user])

	const handleMarkAllAsRead = async () => {
		if (!user?.id) return
		try {
			await notificationsAPI.markAllAsRead(user.id)
			loadNotifications()
		} catch (err) {
			console.error('Failed to mark all as read:', err)
		}
	}

	const handleMarkAsRead = async (id: string) => {
		try {
			await notificationsAPI.markAsRead(id)
			loadNotifications()
		} catch (err) {
			console.error('Failed to mark notification as read:', err)
		}
	}

	const unreadCount = notifications.filter(n => !n.is_read).length

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
					<div style={S.logoIcon}>
						<Home size={18} />
					</div>
					<span style={S.logoText}>DormOS</span>
				</div>

				{/* Nav */}
				<nav style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '1px' }}>
					{navItems.map(item => (
						<NavLink key={item.to} to={item.to} style={({ isActive }) => navLinkStyle(isActive)}>
							<item.icon size={16} style={{ opacity: 0.8 }} />
							{item.label}
						</NavLink>
					))}

					{(user?.role === 'admin' || user?.role === 'manager') && (
						<>
							<div style={S.divider} />
							<NavLink to="/admin" style={({ isActive }) => navLinkStyle(isActive)}>
								<Settings size={16} style={{ opacity: 0.8 }} />
								Admin Panel
							</NavLink>
						</>
					)}
				</nav>

				{/* Notifications Bell */}
				<div style={{ padding: '4px 10px', marginTop: '12px', position: 'relative' }}>
					<button
						onClick={() => setShowNotifications(!showNotifications)}
						style={{
							width: '100%',
							padding: '8px 12px',
							background: showNotifications ? 'rgba(255,255,255,0.06)' : 'rgba(255,255,255,0.02)',
							border: '1px solid rgba(255,255,255,0.06)',
							borderRadius: '8px',
							color: 'var(--text2)',
							fontSize: '13px',
							fontWeight: 510,
							cursor: 'pointer',
							display: 'flex',
							alignItems: 'center',
							justifyContent: 'space-between',
							gap: '8px',
							transition: 'all 0.15s'
						}}
						onMouseOver={e => e.currentTarget.style.background = 'rgba(255,255,255,0.06)'}
						onMouseOut={e => { if (!showNotifications) e.currentTarget.style.background = 'rgba(255,255,255,0.02)' }}
					>
						<div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
							<Bell size={14} style={{ color: unreadCount > 0 ? 'var(--accent)' : 'var(--text3)' }} />
							<span>Notifications</span>
						</div>
						{unreadCount > 0 && (
							<span style={{
								background: 'var(--accent)',
								color: '#fff',
								fontSize: '10px',
								fontWeight: 600,
								padding: '1px 6px',
								borderRadius: '9999px',
								minWidth: '16px',
								textAlign: 'center'
							}}>
								{unreadCount}
							</span>
						)}
					</button>

					{/* Glassmorphic Notifications Popup */}
					{showNotifications && (
						<div style={{
							position: 'absolute',
							bottom: '100%',
							left: '10px',
							width: '320px',
							maxHeight: '400px',
							background: 'rgba(30, 30, 45, 0.95)',
							backdropFilter: 'blur(16px)',
							WebkitBackdropFilter: 'blur(16px)',
							border: '1px solid rgba(255,255,255,0.08)',
							borderRadius: '12px',
							boxShadow: '0 12px 40px rgba(0,0,0,0.5)',
							zIndex: 100,
							display: 'flex',
							flexDirection: 'column',
							marginBottom: '8px',
							overflow: 'hidden'
						}}>
							<div style={{
								display: 'flex',
								justifyContent: 'space-between',
								alignItems: 'center',
								padding: '12px 16px',
								borderBottom: '1px solid rgba(255,255,255,0.06)',
								background: 'rgba(255,255,255,0.01)'
							}}>
								<span style={{ fontSize: '13px', fontWeight: 600, color: 'var(--text)' }}>Dormitory Notifications</span>
								{unreadCount > 0 && (
									<button
										onClick={handleMarkAllAsRead}
										style={{
											background: 'none',
											border: 'none',
											color: 'var(--accent)',
											fontSize: '11px',
											fontWeight: 500,
											cursor: 'pointer',
											display: 'flex',
											alignItems: 'center',
											gap: '4px'
										}}
									>
										<CheckCheck size={12} /> Mark all read
									</button>
								)}
							</div>

							<div style={{ overflowY: 'auto', flex: 1, display: 'flex', flexDirection: 'column' }}>
								{notifications.length === 0 ? (
									<div style={{ padding: '32px 16px', textAlign: 'center', color: 'var(--text4)', fontSize: '12px' }}>
										You have no active notifications.
									</div>
								) : (
									notifications.map((notif) => (
										<div
											key={notif.notification_id}
											onClick={() => !notif.is_read && handleMarkAsRead(notif.notification_id)}
											style={{
												padding: '12px 16px',
												borderBottom: '1px solid rgba(255,255,255,0.04)',
												cursor: notif.is_read ? 'default' : 'pointer',
												background: notif.is_read ? 'transparent' : 'rgba(113,112,255,0.03)',
												transition: 'background 0.15s',
												textAlign: 'left'
											}}
											onMouseOver={e => { if (!notif.is_read) e.currentTarget.style.background = 'rgba(113,112,255,0.06)' }}
											onMouseOut={e => { if (!notif.is_read) e.currentTarget.style.background = 'rgba(113,112,255,0.03)' }}
										>
											<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '4px' }}>
												<span style={{ fontSize: '12.5px', fontWeight: notif.is_read ? 500 : 600, color: notif.is_read ? 'var(--text3)' : 'var(--text)' }}>
													{notif.title}
												</span>
												{!notif.is_read && (
													<span style={{ width: '6px', height: '6px', background: 'var(--accent)', borderRadius: '50%', flexShrink: 0, marginLeft: '6px', marginTop: '4px' }} />
												)}
											</div>
											<p style={{ fontSize: '11.5px', color: 'var(--text4)', margin: 0, lineHeight: 1.4 }}>
												{notif.message}
											</p>
											<span style={{ fontSize: '9px', color: 'rgba(255,255,255,0.15)', marginTop: '6px', display: 'block' }}>
												{new Date(notif.created_at * 1000).toLocaleString()}
											</span>
										</div>
									))
								)}
							</div>
						</div>
					)}
				</div>

				{/* User */}
				<div style={S.user}>
					<div style={S.userInfo}>
						<div style={S.avatar}>{user?.name?.[0]?.toUpperCase() || 'U'}</div>
						<div style={{ flex: 1, minWidth: 0 }}>
							<div style={{ fontSize: '13px', fontWeight: 510, color: 'var(--text)', letterSpacing: '-0.13px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
								{user?.name || 'User'}
							</div>
							<div style={{ fontSize: '11px', color: 'var(--text4)' }}>
								{user?.role ? user.role.toUpperCase() : 'Student'} {user?.room_number ? `• Rm ${user.room_number}` : ''}
							</div>
						</div>
					</div>
					<button
						onClick={() => { logout(); navigate('/login') }}
						style={S.logoutBtn}
						onMouseOver={e => { (e.currentTarget.style.color = 'var(--text2)'); (e.currentTarget.style.background = 'rgba(255,255,255,0.03)') }}
						onMouseOut={e => { (e.currentTarget.style.color = 'var(--text4)'); (e.currentTarget.style.background = 'transparent') }}
					>
						<LogOut size={14} />
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
