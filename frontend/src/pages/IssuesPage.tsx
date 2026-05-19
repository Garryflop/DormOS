import { useEffect, useState } from 'react'
import { issuesAPI } from '../api'
import type { Issue, Category, CreateIssueDto } from '../api'
import { useAuth } from '../AuthContext'

const STATUS_COLOR: Record<string, string> = {
    open: '#7170ff', in_progress: '#f59e0b', resolved: '#10b981', closed: '#62666d',
}
const STATUS_LABEL: Record<string, string> = {
    open: 'Open', in_progress: 'In Progress', resolved: 'Resolved', closed: 'Closed',
}

const inputStyle = {
    width: '100%', padding: '8px 12px',
    background: 'rgba(255,255,255,0.02)',
    border: '1px solid rgba(255,255,255,0.08)',
    borderRadius: '6px', color: 'var(--text2)',
    fontSize: '14px', outline: 'none',
    transition: 'border-color 0.15s',
}

export default function IssuesPage() {
    const { user } = useAuth()
    const [issues, setIssues] = useState<Issue[]>([])
    const [categories, setCategories] = useState<Category[]>([])
    const [loading, setLoading] = useState(true)
    const [showForm, setShowForm] = useState(false)
    const [selected, setSelected] = useState<Issue | null>(null)
    const [comments, setComments] = useState<any[]>([])
    const [comment, setComment] = useState('')
    const [filter, setFilter] = useState('all')
    const [form, setForm] = useState<CreateIssueDto>({
        room_number: user?.room_number || '', category_id: '', title: '', description: '',
    })
    const [submitting, setSubmitting] = useState(false)

    const loadIssues = async () => {
        const res = await (user?.role === 'student' ? issuesAPI.listMy() : issuesAPI.listAll())
        setIssues(res.data.issues || [])
    }

    useEffect(() => {
        Promise.all([loadIssues(), issuesAPI.listCategories()])
            .then(([, catsRes]) => setCategories(catsRes.data.categories || []))
            .finally(() => setLoading(false))
    }, [user])

    const openIssue = async (issue: Issue) => {
        setSelected(issue)
        const res = await issuesAPI.listComments(issue.id)
        setComments(res.data.comments || [])
    }

    const handleCreate = async (e: React.FormEvent) => {
        e.preventDefault()
        setSubmitting(true)
        try {
            await issuesAPI.create(form)
            await loadIssues()
            setShowForm(false)
            setForm({ room_number: user?.room_number || '', category_id: '', title: '', description: '' })
        } finally { setSubmitting(false) }
    }

    const handleComment = async () => {
        if (!selected || !comment.trim()) return
        await issuesAPI.addComment(selected.id, comment)
        const res = await issuesAPI.listComments(selected.id)
        setComments(res.data.comments || [])
        setComment('')
    }

    const filtered = filter === 'all' ? issues : issues.filter(i => i.status === filter)

    return (
        <div style={{ display: 'flex', gap: '24px', height: 'calc(100vh - 64px)', overflow: 'hidden' }}>
            {/* Left */}
            <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '16px', overflowY: 'auto', paddingRight: '4px' }}>
                {/* Header */}
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexShrink: 0 }}>
                    <div>
                        <h1 style={{ fontSize: '20px', fontWeight: 510, letterSpacing: '-0.24px', marginBottom: '2px' }}>Issues</h1>
                        <p style={{ fontSize: '13px', color: 'var(--text4)' }}>{issues.length} total</p>
                    </div>
                    <button
                        onClick={() => setShowForm(!showForm)}
                        style={{
                            padding: '6px 12px', background: 'var(--accent)',
                            border: 'none', borderRadius: '6px',
                            color: '#fff', fontSize: '13px', fontWeight: 510,
                            cursor: 'pointer', letterSpacing: '-0.13px',
                        }}
                    >
                        New issue
                    </button>
                </div>

                {/* Filters */}
                <div style={{ display: 'flex', gap: '4px', flexShrink: 0 }}>
                    {['all', 'open', 'in_progress', 'resolved'].map(f => (
                        <button
                            key={f}
                            onClick={() => setFilter(f)}
                            style={{
                                padding: '4px 10px', borderRadius: '9999px',
                                border: `1px solid ${filter === f ? 'rgba(255,255,255,0.15)' : 'rgba(255,255,255,0.06)'}`,
                                background: filter === f ? 'rgba(255,255,255,0.05)' : 'transparent',
                                color: filter === f ? 'var(--text2)' : 'var(--text4)',
                                fontSize: '12px', fontWeight: 510, cursor: 'pointer',
                                letterSpacing: '-0.13px',
                            }}
                        >
                            {f === 'all' ? 'All' : STATUS_LABEL[f]}
                        </button>
                    ))}
                </div>

                {/* Create form */}
                {showForm && (
                    <div style={{
                        background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(113,112,255,0.3)',
                        borderRadius: '8px', padding: '20px', flexShrink: 0,
                    }}>
                        <h3 style={{ fontSize: '14px', fontWeight: 510, marginBottom: '14px', letterSpacing: '-0.182px' }}>New issue</h3>
                        <form onSubmit={handleCreate} style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                            <input style={inputStyle} placeholder="Title" value={form.title}
                                   onChange={e => setForm({ ...form, title: e.target.value })} required
                                   onFocus={e => e.target.style.borderColor = 'rgba(113,112,255,0.4)'}
                                   onBlur={e => e.target.style.borderColor = 'rgba(255,255,255,0.08)'}
                            />
                            <textarea style={{ ...inputStyle, minHeight: '72px', resize: 'vertical' as const }}
                                      placeholder="Describe the problem..."
                                      value={form.description} onChange={e => setForm({ ...form, description: e.target.value })} required
                                      onFocus={e => e.target.style.borderColor = 'rgba(113,112,255,0.4)'}
                                      onBlur={e => e.target.style.borderColor = 'rgba(255,255,255,0.08)'}
                            />
                            <div style={{ display: 'flex', gap: '8px' }}>
                                <select style={{ ...inputStyle, flex: 1 }} value={form.category_id}
                                        onChange={e => setForm({ ...form, category_id: e.target.value })} required>
                                    <option value="">Category</option>
                                    {categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
                                </select>
                                <input style={{ ...inputStyle, flex: 1 }} placeholder="Room"
                                       value={form.room_number} onChange={e => setForm({ ...form, room_number: e.target.value })} required
                                       onFocus={e => e.target.style.borderColor = 'rgba(113,112,255,0.4)'}
                                       onBlur={e => e.target.style.borderColor = 'rgba(255,255,255,0.08)'}
                                />
                            </div>
                            <div style={{ display: 'flex', gap: '8px' }}>
                                <button type="submit" disabled={submitting} style={{
                                    padding: '7px 16px', background: submitting ? 'rgba(94,106,210,0.5)' : 'var(--accent)',
                                    border: 'none', borderRadius: '6px', color: '#fff',
                                    fontSize: '13px', fontWeight: 510, cursor: 'pointer',
                                }}>
                                    {submitting ? 'Submitting...' : 'Submit'}
                                </button>
                                <button type="button" onClick={() => setShowForm(false)} style={{
                                    padding: '7px 14px', background: 'transparent',
                                    border: '1px solid rgba(255,255,255,0.08)', borderRadius: '6px',
                                    color: 'var(--text3)', fontSize: '13px', cursor: 'pointer',
                                }}>
                                    Cancel
                                </button>
                            </div>
                        </form>
                    </div>
                )}

                {/* Issues list */}
                {loading ? (
                    <div style={{ color: 'var(--text4)', fontSize: '13px', padding: '32px 0', textAlign: 'center' }}>Loading...</div>
                ) : filtered.length === 0 ? (
                    <div style={{
                        background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(255,255,255,0.08)',
                        borderRadius: '8px', padding: '40px', textAlign: 'center', color: 'var(--text4)', fontSize: '14px',
                    }}>
                        No issues found
                    </div>
                ) : (
                    <div style={{
                        background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(255,255,255,0.08)',
                        borderRadius: '8px', overflow: 'hidden',
                    }}>
                        {filtered.map((issue, i) => (
                            <div
                                key={issue.id}
                                onClick={() => openIssue(issue)}
                                style={{
                                    display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                                    padding: '12px 16px',
                                    borderBottom: i < filtered.length - 1 ? '1px solid rgba(255,255,255,0.05)' : 'none',
                                    cursor: 'pointer', transition: 'background 0.1s',
                                    background: selected?.id === issue.id ? 'rgba(113,112,255,0.06)' : 'transparent',
                                }}
                                onMouseOver={e => { if (selected?.id !== issue.id) e.currentTarget.style.background = 'rgba(255,255,255,0.02)' }}
                                onMouseOut={e => { if (selected?.id !== issue.id) e.currentTarget.style.background = 'transparent' }}
                            >
                                <div style={{ flex: 1, minWidth: 0 }}>
                                    <div style={{ fontSize: '14px', fontWeight: 510, letterSpacing: '-0.182px', marginBottom: '2px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                                        {issue.title}
                                    </div>
                                    <div style={{ fontSize: '12px', color: 'var(--text4)' }}>
                                        Room {issue.room_number} · {new Date(issue.created_at).toLocaleDateString()}
                                    </div>
                                </div>
                                <span style={{
                                    padding: '2px 8px', marginLeft: '12px',
                                    background: `${STATUS_COLOR[issue.status]}18`,
                                    color: STATUS_COLOR[issue.status],
                                    borderRadius: '9999px', fontSize: '11px', fontWeight: 510,
                                    whiteSpace: 'nowrap',
                                }}>
                  {STATUS_LABEL[issue.status]}
                </span>
                            </div>
                        ))}
                    </div>
                )}
            </div>

            {/* Right panel */}
            {selected && (
                <div style={{
                    width: '340px', flexShrink: 0,
                    background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(255,255,255,0.08)',
                    borderRadius: '8px', padding: '20px',
                    display: 'flex', flexDirection: 'column', gap: '14px',
                    overflowY: 'auto', height: 'fit-content', maxHeight: '100%',
                }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                        <h2 style={{ fontSize: '15px', fontWeight: 510, letterSpacing: '-0.165px', flex: 1, paddingRight: '8px' }}>
                            {selected.title}
                        </h2>
                        <button onClick={() => setSelected(null)} style={{
                            background: 'none', border: 'none', color: 'var(--text4)', cursor: 'pointer', fontSize: '16px', lineHeight: 1,
                        }}>×</button>
                    </div>

                    <span style={{
                        display: 'inline-block', padding: '2px 8px', width: 'fit-content',
                        background: `${STATUS_COLOR[selected.status]}18`, color: STATUS_COLOR[selected.status],
                        borderRadius: '9999px', fontSize: '11px', fontWeight: 510,
                    }}>
            {STATUS_LABEL[selected.status]}
          </span>

                    <p style={{ fontSize: '13px', color: 'var(--text3)', lineHeight: 1.6 }}>{selected.description}</p>

                    <div style={{ fontSize: '12px', color: 'var(--text4)', display: 'flex', flexDirection: 'column', gap: '4px' }}>
                        <span>Room: <span style={{ color: 'var(--text3)' }}>{selected.room_number}</span></span>
                        <span>Created: <span style={{ color: 'var(--text3)' }}>{new Date(selected.created_at).toLocaleString()}</span></span>
                    </div>

                    <div style={{ borderTop: '1px solid rgba(255,255,255,0.05)', paddingTop: '14px' }}>
                        <h4 style={{ fontSize: '12px', fontWeight: 510, color: 'var(--text3)', marginBottom: '10px', letterSpacing: '0.02em', textTransform: 'uppercase' }}>
                            Comments {comments.length > 0 && `(${comments.length})`}
                        </h4>
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginBottom: '12px' }}>
                            {comments.length === 0 ? (
                                <div style={{ fontSize: '12px', color: 'var(--text4)' }}>No comments yet</div>
                            ) : comments.map((c: any) => (
                                <div key={c.id} style={{
                                    background: 'rgba(255,255,255,0.03)', borderRadius: '6px', padding: '8px 10px',
                                }}>
                                    <div style={{ fontSize: '11px', color: 'var(--text4)', marginBottom: '3px' }}>
                                        {new Date(c.created_at).toLocaleString()}
                                    </div>
                                    <div style={{ fontSize: '13px', color: 'var(--text2)' }}>{c.text}</div>
                                </div>
                            ))}
                        </div>
                        <div style={{ display: 'flex', gap: '6px' }}>
                            <input
                                style={{ ...inputStyle, flex: 1, fontSize: '13px', padding: '6px 10px' }}
                                placeholder="Add a comment..."
                                value={comment}
                                onChange={e => setComment(e.target.value)}
                                onKeyDown={e => e.key === 'Enter' && handleComment()}
                            />
                            <button onClick={handleComment} style={{
                                padding: '6px 12px', background: 'var(--accent)',
                                border: 'none', borderRadius: '6px',
                                color: '#fff', fontSize: '12px', cursor: 'pointer',
                            }}>
                                Send
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}
