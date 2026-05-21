import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import { issuesAPI } from '../api'
import type { Issue } from '../api'

const STATUS_COLOR: Record<string, string> = {
    open: '#7170ff', in_progress: '#f59e0b', resolved: '#10b981', closed: '#62666d',
}
const STATUS_LABEL: Record<string, string> = {
    open: 'Open', in_progress: 'In Progress', resolved: 'Resolved', closed: 'Closed',
}

function StatCard({ label, value, color }: { label: string; value: number; color: string }) {
    return (
        <div style={{
            background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(255,255,255,0.08)',
            borderRadius: '8px', padding: '20px 24px',
        }}>
            <div style={{ fontSize: '28px', fontWeight: 510, color, letterSpacing: '-0.704px', marginBottom: '4px' }}>
                {value}
            </div>
            <div style={{ fontSize: '13px', color: 'var(--text3)', letterSpacing: '-0.13px' }}>{label}</div>
        </div>
    )
}

export default function DashboardPage() {
    const { user } = useAuth()
    const navigate = useNavigate()
    const [issues, setIssues] = useState<Issue[]>([])
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        issuesAPI.listMy()
            .then(res => setIssues(res.data.issues || []))
            .catch(() => setIssues([]))
            .finally(() => setLoading(false))
    }, [])

    const stats = {
        total: issues.length,
        open: issues.filter(i => i.status === 'open').length,
        in_progress: issues.filter(i => i.status === 'in_progress').length,
        resolved: issues.filter(i => i.status === 'resolved').length,
    }

    return (
        <div style={{ maxWidth: '900px' }}>
            {/* Header */}
            <div style={{ marginBottom: '32px' }}>
                <h1 style={{ fontSize: '24px', fontWeight: 510, color: 'var(--text)', letterSpacing: '-0.288px', marginBottom: '4px' }}>
                    Good to see you, {user?.name?.split(' ')[0]}
                </h1>
                <p style={{ fontSize: '14px', color: 'var(--text3)', letterSpacing: '-0.13px' }}>
                    Room {user?.room_number} · Overview of your dormitory activity
                </p>
            </div>

            {/* Stats */}
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '12px', marginBottom: '32px' }}>
                <StatCard label="Total Issues" value={loading ? 0 : stats.total} color="var(--text)" />
                <StatCard label="Open" value={loading ? 0 : stats.open} color="#7170ff" />
                <StatCard label="In Progress" value={loading ? 0 : stats.in_progress} color="#f59e0b" />
                <StatCard label="Resolved" value={loading ? 0 : stats.resolved} color="#10b981" />
            </div>

            {/* Recent issues */}
            <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                    <h2 style={{ fontSize: '15px', fontWeight: 510, color: 'var(--text2)', letterSpacing: '-0.165px' }}>
                        Recent Issues
                    </h2>
                    <button
                        onClick={() => navigate('/issues')}
                        style={{
                            padding: '4px 10px', background: 'transparent',
                            border: '1px solid rgba(255,255,255,0.08)', borderRadius: '6px',
                            color: 'var(--text3)', fontSize: '12px', fontWeight: 510, cursor: 'pointer',
                            letterSpacing: '-0.13px',
                        }}
                    >
                        View all
                    </button>
                </div>

                <div style={{
                    background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(255,255,255,0.08)',
                    borderRadius: '8px', overflow: 'hidden',
                }}>
                    {loading ? (
                        <div style={{ padding: '32px', textAlign: 'center', color: 'var(--text4)', fontSize: '13px' }}>Loading...</div>
                    ) : issues.length === 0 ? (
                        <div style={{ padding: '40px', textAlign: 'center', color: 'var(--text4)', fontSize: '14px' }}>
                            No issues yet — everything looks good ✓
                        </div>
                    ) : (
                        issues.slice(0, 6).map((issue, i) => (
                            <div
                                key={issue.id}
                                onClick={() => navigate('/issues')}
                                style={{
                                    display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                                    padding: '12px 16px',
                                    borderBottom: i < Math.min(issues.length, 6) - 1 ? '1px solid rgba(255,255,255,0.05)' : 'none',
                                    cursor: 'pointer', transition: 'background 0.1s',
                                }}
                                onMouseOver={e => (e.currentTarget.style.background = 'rgba(255,255,255,0.02)')}
                                onMouseOut={e => (e.currentTarget.style.background = 'transparent')}
                            >
                                <div style={{ flex: 1, minWidth: 0 }}>
                                    <div style={{ fontSize: '14px', fontWeight: 510, color: 'var(--text)', letterSpacing: '-0.182px', marginBottom: '2px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                                        {issue.title}
                                    </div>
                                    <div style={{ fontSize: '12px', color: 'var(--text4)' }}>
                                        Room {issue.room_number} · {new Date(issue.created_at).toLocaleDateString()}
                                    </div>
                                </div>
                                <span style={{
                                    padding: '2px 8px', marginLeft: '16px',
                                    background: `${STATUS_COLOR[issue.status]}18`,
                                    color: STATUS_COLOR[issue.status],
                                    borderRadius: '9999px', fontSize: '11px', fontWeight: 510,
                                    whiteSpace: 'nowrap', letterSpacing: '-0.13px',
                                }}>
                  {STATUS_LABEL[issue.status]}
                </span>
                            </div>
                        ))
                    )}
                </div>
            </div>
        </div>
    )
}
