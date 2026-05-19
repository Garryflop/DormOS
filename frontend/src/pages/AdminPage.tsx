import { useEffect, useState } from 'react'
import { issuesAPI } from '../api'
import type { Issue } from '../api'

const STATUS_COLOR: Record<string, string> = {
    open: '#7170ff', in_progress: '#f59e0b', resolved: '#10b981', closed: '#62666d',
}

export default function AdminPage() {
    const [issues, setIssues] = useState<Issue[]>([])
    const [loading, setLoading] = useState(true)

    const load = () => issuesAPI.listAll()
        .then(res => setIssues(res.data.issues || []))
        .finally(() => setLoading(false))

    useEffect(() => { load() }, [])

    const updateStatus = async (id: string, status: string) => {
        await issuesAPI.updateStatus(id, status)
        load()
    }

    return (
        <div style={{ maxWidth: '1000px' }}>
            <div style={{ marginBottom: '28px' }}>
                <h1 style={{ fontSize: '20px', fontWeight: 510, letterSpacing: '-0.24px', marginBottom: '4px' }}>Admin</h1>
                <p style={{ fontSize: '13px', color: 'var(--text4)' }}>Manage all issues across the dormitory</p>
            </div>

            {loading ? (
                <div style={{ color: 'var(--text4)', fontSize: '13px', textAlign: 'center', padding: '48px' }}>Loading...</div>
            ) : (
                <div style={{
                    background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(255,255,255,0.08)',
                    borderRadius: '8px', overflow: 'hidden',
                }}>
                    <div style={{
                        display: 'grid', gridTemplateColumns: '1fr 80px 120px 100px 130px',
                        padding: '10px 16px',
                        borderBottom: '1px solid rgba(255,255,255,0.05)',
                    }}>
                        {['Issue', 'Room', 'Status', 'Created', 'Action'].map(h => (
                            <div key={h} style={{ fontSize: '11px', fontWeight: 510, color: 'var(--text4)', letterSpacing: '0.05em', textTransform: 'uppercase' }}>
                                {h}
                            </div>
                        ))}
                    </div>

                    {issues.length === 0 ? (
                        <div style={{ padding: '40px', textAlign: 'center', color: 'var(--text4)', fontSize: '13px' }}>
                            No issues found
                        </div>
                    ) : issues.map((issue, i) => (
                        <div
                            key={issue.id}
                            style={{
                                display: 'grid', gridTemplateColumns: '1fr 80px 120px 100px 130px',
                                padding: '12px 16px', alignItems: 'center',
                                borderBottom: i < issues.length - 1 ? '1px solid rgba(255,255,255,0.05)' : 'none',
                                transition: 'background 0.1s',
                            }}
                            onMouseOver={e => (e.currentTarget.style.background = 'rgba(255,255,255,0.02)')}
                            onMouseOut={e => (e.currentTarget.style.background = 'transparent')}
                        >
                            <div style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', paddingRight: '12px' }}>
                                <div style={{ fontSize: '14px', fontWeight: 510, letterSpacing: '-0.182px', color: 'var(--text)' }}>
                                    {issue.title}
                                </div>
                            </div>
                            <div style={{ fontSize: '13px', color: 'var(--text3)' }}>{issue.room_number}</div>
                            <span style={{
                                display: 'inline-block', padding: '2px 8px',
                                background: `${STATUS_COLOR[issue.status]}18`,
                                color: STATUS_COLOR[issue.status],
                                borderRadius: '9999px', fontSize: '11px', fontWeight: 510,
                                width: 'fit-content',
                            }}>
                {issue.status.replace('_', ' ')}
              </span>
                            <div style={{ fontSize: '12px', color: 'var(--text4)' }}>
                                {new Date(issue.created_at).toLocaleDateString()}
                            </div>
                            <select
                                value={issue.status}
                                onChange={e => updateStatus(issue.id, e.target.value)}
                                style={{
                                    padding: '5px 8px',
                                    background: 'rgba(255,255,255,0.04)',
                                    border: '1px solid rgba(255,255,255,0.08)',
                                    borderRadius: '6px', color: 'var(--text2)',
                                    fontSize: '12px', cursor: 'pointer', outline: 'none',
                                }}
                            >
                                <option value="open">Open</option>
                                <option value="in_progress">In Progress</option>
                                <option value="resolved">Resolved</option>
                                <option value="closed">Closed</option>
                            </select>
                        </div>
                    ))}
                </div>
            )}
        </div>
    )
}
