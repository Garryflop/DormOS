import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import { issuesAPI } from '../api'
import type { Issue } from '../api'
import { 
	CheckCircle2, 
	Clock, 
	AlertCircle, 
	Calendar, 
	MapPin, 
	ArrowRight,
	Activity
} from 'lucide-react'

const STATUS_COLOR: Record<string, string> = {
	open: '#7170ff', in_progress: '#f59e0b', resolved: '#10b981', closed: '#62666d',
}
const STATUS_LABEL: Record<string, string> = {
	open: 'Open', in_progress: 'In Progress', resolved: 'Resolved', closed: 'Closed',
}

interface EventItem {
	id: string
	title: string
	time: string
	location: string
	category: string
	color: string
}

const mockEvents: EventItem[] = [
	{ id: '1', title: 'Table Tennis Championship', time: 'Today at 19:00', location: 'Recreation Hall', category: 'Sport', color: '#10b981' },
	{ id: '2', title: 'Community General Meeting', time: 'Saturday at 14:00', location: 'Conference Room', category: 'Meeting', color: '#7170ff' },
	{ id: '3', title: 'Dormitory Clean-up Day', time: 'Sunday at 10:00', location: 'All Floors', category: 'Social', color: '#f59e0b' },
]

function StatCard({ label, value, color, icon: Icon }: { label: string; value: number; color: string; icon: any }) {
	return (
		<div style={{
			background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(255,255,255,0.06)',
			borderRadius: '12px', padding: '20px 24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center',
			transition: 'transform 0.2s, border-color 0.2s', cursor: 'default'
		}}
		onMouseOver={e => e.currentTarget.style.borderColor = 'rgba(255,255,255,0.12)'}
		onMouseOut={e => e.currentTarget.style.borderColor = 'rgba(255,255,255,0.06)'}
		>
			<div>
				<div style={{ fontSize: '28px', fontWeight: 600, color, letterSpacing: '-0.704px', marginBottom: '4px' }}>
					{value}
				</div>
				<div style={{ fontSize: '13px', color: 'var(--text3)', letterSpacing: '-0.13px' }}>{label}</div>
			</div>
			<div style={{ background: `${color}12`, padding: '10px', borderRadius: '10px', color: color }}>
				<Icon size={20} />
			</div>
		</div>
	)
}

export default function DashboardPage() {
	const { user } = useAuth()
	const navigate = useNavigate()
	const [issues, setIssues] = useState<Issue[]>([])
	const [loading, setLoading] = useState(true)

	useEffect(() => {
		const isStaff = user?.role === 'manager' || user?.role === 'admin'
		const fetchIssues = isStaff ? issuesAPI.listAll() : issuesAPI.listMy()
		fetchIssues
			.then(res => setIssues(res.data.issues || []))
			.catch(() => setIssues([]))
			.finally(() => setLoading(false))
	}, [user])

	const stats = {
		total: issues.length,
		open: issues.filter(i => i.status === 'open').length,
		in_progress: issues.filter(i => i.status === 'in_progress').length,
		resolved: issues.filter(i => i.status === 'resolved').length,
	}

	return (
		<div style={{ width: '100%', display: 'flex', flexDirection: 'column', gap: '32px' }}>
			{/* Header */}
			<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
				<div>
					<h1 style={{ fontSize: '24px', fontWeight: 600, color: 'var(--text)', letterSpacing: '-0.288px', marginBottom: '6px' }}>
						Welcome back, {user?.name?.split(' ')[0]}
					</h1>
					<p style={{ fontSize: '14px', color: 'var(--text3)', letterSpacing: '-0.13px', display: 'flex', alignItems: 'center', gap: '6px' }}>
						<span>Room {user?.room_number || 'Not Assigned'}</span> 
						<span style={{ color: 'rgba(255,255,255,0.15)' }}>•</span>
						<span style={{ color: 'var(--accent)', fontWeight: 500 }}>{user?.role?.toUpperCase() || 'STUDENT'}</span>
					</p>
				</div>
				<div style={{ display: 'flex', gap: '8px' }}>
					<button
						onClick={() => navigate('/issues')}
						style={{
							padding: '8px 16px', background: 'var(--accent)',
							border: 'none', borderRadius: '8px',
							color: '#fff', fontSize: '13px', fontWeight: 510, cursor: 'pointer',
							letterSpacing: '-0.13px', display: 'flex', alignItems: 'center', gap: '8px',
							transition: 'opacity 0.15s'
						}}
						onMouseOver={e => e.currentTarget.style.opacity = '0.9'}
						onMouseOut={e => e.currentTarget.style.opacity = '1'}
					>
						New Issue
					</button>
				</div>
			</div>

			{/* Stats */}
			<div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '16px' }}>
				<StatCard label="Total Issues" value={loading ? 0 : stats.total} color="var(--text)" icon={Activity} />
				<StatCard label="Open" value={loading ? 0 : stats.open} color="#7170ff" icon={AlertCircle} />
				<StatCard label="In Progress" value={loading ? 0 : stats.in_progress} color="#f59e0b" icon={Clock} />
				<StatCard label="Resolved" value={loading ? 0 : stats.resolved} color="#10b981" icon={CheckCircle2} />
			</div>

			{/* Two Column Section: Issues & Events */}
			<div style={{ display: 'grid', gridTemplateColumns: '1fr 360px', gap: '24px' }}>
				{/* Recent issues */}
				<div>
					<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
						<h2 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--text2)', letterSpacing: '-0.165px' }}>
							Your Recent Issues
						</h2>
						<button
							onClick={() => navigate('/issues')}
							style={{
								padding: '6px 12px', background: 'transparent',
								border: '1px solid rgba(255,255,255,0.08)', borderRadius: '6px',
								color: 'var(--text3)', fontSize: '12px', fontWeight: 510, cursor: 'pointer',
								letterSpacing: '-0.13px', display: 'flex', alignItems: 'center', gap: '4px'
							}}
						>
							View all <ArrowRight size={12} />
						</button>
					</div>

					<div style={{
						background: 'rgba(255,255,255,0.01)', border: '1px solid rgba(255,255,255,0.06)',
						borderRadius: '12px', overflow: 'hidden',
					}}>
						{loading ? (
							<div style={{ padding: '32px', textAlign: 'center', color: 'var(--text4)', fontSize: '13px' }}>Loading...</div>
						) : issues.length === 0 ? (
							<div style={{ padding: '48px 24px', textAlign: 'center', color: 'var(--text4)', fontSize: '14px' }}>
								No issues filed yet — everything is in order.
							</div>
						) : (
							issues.slice(0, 5).map((issue, i) => (
								<div
									key={issue.id}
									onClick={() => navigate('/issues')}
									style={{
										display: 'flex', alignItems: 'center', justifyContent: 'space-between',
										padding: '14px 18px',
										borderBottom: i < Math.min(issues.length, 5) - 1 ? '1px solid rgba(255,255,255,0.04)' : 'none',
										cursor: 'pointer', transition: 'background 0.15s',
									}}
									onMouseOver={e => (e.currentTarget.style.background = 'rgba(255,255,255,0.02)')}
									onMouseOut={e => (e.currentTarget.style.background = 'transparent')}
								>
									<div style={{ flex: 1, minWidth: 0 }}>
										<div style={{ fontSize: '14px', fontWeight: 510, color: 'var(--text)', letterSpacing: '-0.182px', marginBottom: '4px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
											{issue.title}
										</div>
										<div style={{ fontSize: '12px', color: 'var(--text4)', display: 'flex', alignItems: 'center', gap: '4px' }}>
											<span>Room {issue.room_number}</span>
											<span>•</span>
											<span>{new Date(issue.created_at).toLocaleDateString()}</span>
										</div>
									</div>
									<span style={{
										padding: '3px 10px', marginLeft: '16px',
										background: `${STATUS_COLOR[issue.status]}12`,
										color: STATUS_COLOR[issue.status],
										borderRadius: '9999px', fontSize: '11px', fontWeight: 600,
										whiteSpace: 'nowrap', letterSpacing: '-0.13px', textTransform: 'uppercase'
									}}>
										{STATUS_LABEL[issue.status]}
									</span>
								</div>
							))
						)}
					</div>
				</div>

				{/* Active Events */}
				<div>
					<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
						<h2 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--text2)', letterSpacing: '-0.165px' }}>
							Dormitory Events
						</h2>
						<button
							onClick={() => navigate('/activities')}
							style={{
								padding: '6px 12px', background: 'transparent',
								border: '1px solid rgba(255,255,255,0.08)', borderRadius: '6px',
								color: 'var(--text3)', fontSize: '12px', fontWeight: 510, cursor: 'pointer',
								letterSpacing: '-0.13px', display: 'flex', alignItems: 'center', gap: '4px'
							}}
						>
							Calendar <Calendar size={12} />
						</button>
					</div>

					<div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
						{mockEvents.map(evt => (
							<div key={evt.id} style={{
								background: 'rgba(255,255,255,0.01)', border: '1px solid rgba(255,255,255,0.06)',
								borderRadius: '12px', padding: '16px', display: 'flex', flexDirection: 'column', gap: '10px',
								transition: 'border-color 0.2s', cursor: 'default'
							}}
							onMouseOver={e => e.currentTarget.style.borderColor = 'rgba(255,255,255,0.12)'}
							onMouseOut={e => e.currentTarget.style.borderColor = 'rgba(255,255,255,0.06)'}
							>
								<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: '8px' }}>
									<h3 style={{ fontSize: '14px', fontWeight: 600, color: 'var(--text)', letterSpacing: '-0.154px', margin: 0 }}>
										{evt.title}
									</h3>
									<span style={{
										padding: '2px 8px', borderRadius: '4px', fontSize: '10px', fontWeight: 600,
										background: `${evt.color}12`, color: evt.color, textTransform: 'uppercase'
									}}>
										{evt.category}
									</span>
								</div>
								
								<div style={{ display: 'flex', flexDirection: 'column', gap: '4px', fontSize: '12px', color: 'var(--text4)' }}>
									<div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
										<Calendar size={12} />
										<span>{evt.time}</span>
									</div>
									<div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
										<MapPin size={12} />
										<span>{evt.location}</span>
									</div>
								</div>
							</div>
						))}
					</div>
				</div>
			</div>
		</div>
	)
}
