import { useEffect, useState } from 'react'
import { issuesAPI, filesAPI } from '../api'
import type { Issue, Category, CreateIssueDto, Worker } from '../api'
import { useAuth } from '../AuthContext'
import { 
	Wrench, 
	Plus, 
	X, 
	Send, 
	Calendar, 
	MapPin, 
	Camera, 
	CheckCircle2, 
	Clock, 
	AlertCircle,
	MessageSquare,
	Trash2
} from 'lucide-react'

const STATUS_COLOR: Record<string, string> = {
	open: '#7170ff', in_progress: '#f59e0b', resolved: '#10b981', closed: '#62666d',
}
const STATUS_LABEL: Record<string, string> = {
	open: 'Open', in_progress: 'In Progress', resolved: 'Resolved', closed: 'Closed',
}

const inputStyle = {
	width: '100%', padding: '10px 14px',
	background: 'rgba(255,255,255,0.02)',
	border: '1px solid rgba(255,255,255,0.06)',
	borderRadius: '8px', color: 'var(--text2)',
	fontSize: '14px', outline: 'none',
	transition: 'border-color 0.15s',
}

export default function IssuesPage() {
	const { user } = useAuth()
	const [issues, setIssues] = useState<Issue[]>([])
	const [categories, setCategories] = useState<Category[]>([])
	const [workers, setWorkers] = useState<Worker[]>([])
	const [loading, setLoading] = useState(true)
	const [showForm, setShowForm] = useState(false)
	const [selected, setSelected] = useState<Issue | null>(null)
	const [comments, setComments] = useState<any[]>([])
	const [comment, setComment] = useState('')
	const [filter, setFilter] = useState('all')
	const [form, setForm] = useState<CreateIssueDto>({
		room_number: user?.room_number || '', category_id: '', title: '', description: '', photo_url: '',
	})
	const [uploading, setUploading] = useState(false)
	const [submitting, setSubmitting] = useState(false)

	const isStaff = user?.role === 'manager' || user?.role === 'admin' || user?.role === 'dorm_admin'

	const loadIssues = async () => {
		const res = await (isStaff ? issuesAPI.listAll() : issuesAPI.listMy())
		setIssues(res.data.issues || [])
	}

	const loadInitialData = async () => {
		setLoading(true)
		try {
			await loadIssues()
			const catsRes = await issuesAPI.listCategories()
			setCategories(catsRes.data.categories || [])
			if (isStaff) {
				const workersRes = await issuesAPI.listWorkers()
				setWorkers(workersRes.data.workers || [])
			}
		} catch (err) {
			console.error('Failed to load initial issues data:', err)
		} finally {
			setLoading(false)
		}
	}

	useEffect(() => {
		loadInitialData()
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
			setForm({ room_number: user?.room_number || '', category_id: '', title: '', description: '', photo_url: '' })
		} catch (err) {
			console.error('Failed to create issue:', err)
		} finally { setSubmitting(false) }
	}

	const handleComment = async () => {
		if (!selected || !comment.trim()) return
		await issuesAPI.addComment(selected.id, comment)
		const res = await issuesAPI.listComments(selected.id)
		setComments(res.data.comments || [])
		setComment('')
	}

	const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
		const file = e.target.files?.[0]
		if (!file) return
		setUploading(true)
		try {
			// 1. Upload file using File Service
			const uploadRes = await filesAPI.upload(file)
			const fileId = uploadRes.data.file_id
			
			// 2. Get permanent presigned URL or direct link
			const urlRes = await filesAPI.getUrl(fileId)
			setForm(prev => ({ ...prev, photo_url: urlRes.data.url }))
		} catch (err) {
			console.error('Failed to upload photo:', err)
			alert('Failed to upload photo. Please check File Service.')
		} finally {
			setUploading(false)
		}
	}

	const handleStatusChange = async (status: string) => {
		if (!selected) return
		try {
			await issuesAPI.updateStatus(selected.id, status)
			// Reload issues and update details view
			await loadIssues()
			const updated = issues.find(i => i.id === selected.id)
			if (updated) {
				setSelected({ ...selected, status: status as any })
			} else {
				setSelected(null)
			}
		} catch (err) {
			console.error('Failed to update status:', err)
		}
	}

	const handleAssignWorker = async (workerId: string) => {
		if (!selected) return
		try {
			await issuesAPI.assignWorker(selected.id, workerId)
			await loadIssues()
			setSelected({ ...selected, worker_id: workerId })
			alert('Worker successfully assigned!')
		} catch (err) {
			console.error('Failed to assign worker:', err)
		}
	}

	const handleDeleteIssue = async (id: string) => {
		if (!window.confirm('Are you sure you want to permanently delete this issue?')) return
		try {
			await issuesAPI.delete(id)
			setSelected(null)
			await loadIssues()
		} catch (err) {
			console.error('Failed to delete issue:', err)
		}
	}

	const filtered = filter === 'all' ? issues : issues.filter(i => i.status === filter)

	return (
		<div style={{ display: 'flex', gap: '24px', height: 'calc(100vh - 96px)', overflow: 'hidden' }}>
			{/* Left Column */}
			<div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '16px', overflowY: 'auto', paddingRight: '4px' }}>
				{/* Header */}
				<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexShrink: 0 }}>
					<div>
						<h1 style={{ fontSize: '22px', fontWeight: 600, letterSpacing: '-0.24px', marginBottom: '4px', display: 'flex', alignItems: 'center', gap: '8px' }}>
							<Wrench size={22} className="text-accent" />
							Issues Registry
						</h1>
						<p style={{ fontSize: '13px', color: 'var(--text4)' }}>
							{isStaff ? 'All student reports' : 'Your filed complaints'} · {issues.length} total
						</p>
					</div>
					{!isStaff && (
						<button
							onClick={() => setShowForm(!showForm)}
							style={{
								padding: '8px 16px', background: 'var(--accent)',
								border: 'none', borderRadius: '8px',
								color: '#fff', fontSize: '13px', fontWeight: 510,
								cursor: 'pointer', letterSpacing: '-0.13px', display: 'flex', alignItems: 'center', gap: '6px'
							}}
						>
							<Plus size={16} /> New Issue
						</button>
					)}
				</div>

				{/* Filters */}
				<div style={{ display: 'flex', gap: '6px', flexShrink: 0 }}>
					{['all', 'open', 'in_progress', 'resolved'].map(f => (
						<button
							key={f}
							onClick={() => setFilter(f)}
							style={{
								padding: '6px 14px', borderRadius: '9999px',
								border: `1px solid ${filter === f ? 'rgba(255,255,255,0.12)' : 'rgba(255,255,255,0.04)'}`,
								background: filter === f ? 'rgba(255,255,255,0.04)' : 'transparent',
								color: filter === f ? 'var(--text2)' : 'var(--text4)',
								fontSize: '12px', fontWeight: 510, cursor: 'pointer',
								letterSpacing: '-0.13px', display: 'flex', alignItems: 'center', gap: '6px'
							}}
						>
							{f === 'open' && <AlertCircle size={12} color="#7170ff" />}
							{f === 'in_progress' && <Clock size={12} color="#f59e0b" />}
							{f === 'resolved' && <CheckCircle2 size={12} color="#10b981" />}
							{f === 'all' ? 'All Issues' : STATUS_LABEL[f]}
						</button>
					))}
				</div>

				{/* Create form */}
				{showForm && !isStaff && (
					<div style={{
						background: 'rgba(255,255,255,0.01)', border: '1px solid rgba(113,112,255,0.2)',
						borderRadius: '12px', padding: '20px', flexShrink: 0, display: 'flex', flexDirection: 'column', gap: '16px'
					}}>
						<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
							<h3 style={{ fontSize: '15px', fontWeight: 600, letterSpacing: '-0.182px' }}>File a New Complaint</h3>
							<button onClick={() => setShowForm(false)} style={{ background: 'none', border: 'none', color: 'var(--text4)', cursor: 'pointer' }}>
								<X size={16} />
							</button>
						</div>
						<form onSubmit={handleCreate} style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
							<input style={inputStyle} placeholder="Title / Short Summary" value={form.title}
								onChange={e => setForm({ ...form, title: e.target.value })} required
							/>
							<textarea style={{ ...inputStyle, minHeight: '90px', resize: 'vertical' as const }}
								placeholder="Provide a detailed description of the issue..."
								value={form.description} onChange={e => setForm({ ...form, description: e.target.value })} required
							/>
							<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
								<select style={inputStyle} value={form.category_id}
									onChange={e => setForm({ ...form, category_id: e.target.value })} required>
									<option value="">Select Category</option>
									{categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
								</select>
								<input style={inputStyle} placeholder="Room number"
									value={form.room_number} onChange={e => setForm({ ...form, room_number: e.target.value })} required
								/>
							</div>

							{/* Photo Upload using File Service */}
							<div style={{
								border: '1px dashed rgba(255,255,255,0.08)', borderRadius: '8px', padding: '14px',
								display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: '8px',
								background: 'rgba(255,255,255,0.005)'
							}}>
								{form.photo_url ? (
									<div style={{ position: 'relative', width: '100%' }}>
										<img src={form.photo_url} alt="Uploaded preview" style={{ width: '100%', maxHeight: '160px', objectFit: 'cover', borderRadius: '6px' }} />
										<button 
											type="button" 
											onClick={() => setForm(prev => ({ ...prev, photo_url: '' }))}
											style={{
												position: 'absolute', top: '8px', right: '8px', background: 'rgba(0,0,0,0.6)', 
												border: 'none', borderRadius: '4px', padding: '4px', color: '#fff', cursor: 'pointer'
											}}
										>
											<Trash2 size={14} />
										</button>
									</div>
								) : (
									<label style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '6px', cursor: 'pointer', color: 'var(--text3)' }}>
										<Camera size={24} className="text-accent" />
										<span style={{ fontSize: '13px', fontWeight: 500 }}>{uploading ? 'Uploading to MinIO...' : 'Upload Problem Photo'}</span>
										<span style={{ fontSize: '11px', color: 'var(--text4)' }}>Supports PNG, JPG (Max 5MB)</span>
										<input type="file" accept="image/*" onChange={handleFileUpload} style={{ display: 'none' }} disabled={uploading} />
									</label>
								)}
							</div>

							<div style={{ display: 'flex', gap: '10px', marginTop: '4px' }}>
								<button type="submit" disabled={submitting || uploading} style={{
									padding: '8px 20px', background: submitting ? 'rgba(94,106,210,0.5)' : 'var(--accent)',
									border: 'none', borderRadius: '8px', color: '#fff',
									fontSize: '13px', fontWeight: 510, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px'
								}}>
									<Send size={14} /> {submitting ? 'Submitting...' : 'Submit Issue'}
								</button>
								<button type="button" onClick={() => setShowForm(false)} style={{
									padding: '8px 16px', background: 'transparent',
									border: '1px solid rgba(255,255,255,0.08)', borderRadius: '8px',
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
					<div style={{ color: 'var(--text4)', fontSize: '13px', padding: '32px 0', textAlign: 'center' }}>Loading issues...</div>
				) : filtered.length === 0 ? (
					<div style={{
						background: 'rgba(255,255,255,0.01)', border: '1px solid rgba(255,255,255,0.06)',
						borderRadius: '12px', padding: '48px', textAlign: 'center', color: 'var(--text4)', fontSize: '14px',
					}}>
						No issues found in this registry.
					</div>
				) : (
					<div style={{
						background: 'rgba(255,255,255,0.01)', border: '1px solid rgba(255,255,255,0.06)',
						borderRadius: '12px', overflow: 'hidden',
					}}>
						{filtered.map((issue, i) => (
							<div
								key={issue.id}
								onClick={() => openIssue(issue)}
								style={{
									display: 'flex', alignItems: 'center', justifyContent: 'space-between',
									padding: '14px 18px',
									borderBottom: i < filtered.length - 1 ? '1px solid rgba(255,255,255,0.04)' : 'none',
									cursor: 'pointer', transition: 'background 0.15s',
									background: selected?.id === issue.id ? 'rgba(113,112,255,0.05)' : 'transparent',
								}}
								onMouseOver={e => { if (selected?.id !== issue.id) e.currentTarget.style.background = 'rgba(255,255,255,0.015)' }}
								onMouseOut={e => { if (selected?.id !== issue.id) e.currentTarget.style.background = 'transparent' }}
							>
								<div style={{ flex: 1, minWidth: 0 }}>
									<div style={{ fontSize: '14px', fontWeight: 510, letterSpacing: '-0.182px', marginBottom: '4px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
										{issue.title}
									</div>
									<div style={{ fontSize: '12px', color: 'var(--text4)', display: 'flex', alignItems: 'center', gap: '6px' }}>
										<span>Room {issue.room_number}</span>
										<span>•</span>
										<span>{new Date(issue.created_at).toLocaleDateString()}</span>
										{issue.photo_url && (
											<>
												<span>•</span>
												<span style={{ color: 'var(--accent)', display: 'flex', alignItems: 'center', gap: '2px' }}>
													<Camera size={11} /> Photo
												</span>
											</>
										)}
									</div>
								</div>
								<div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
									<span style={{
										padding: '3px 10px',
										background: `${STATUS_COLOR[issue.status]}12`,
										color: STATUS_COLOR[issue.status],
										borderRadius: '9999px', fontSize: '11px', fontWeight: 600,
										whiteSpace: 'nowrap', textTransform: 'uppercase'
									}}>
										{STATUS_LABEL[issue.status]}
									</span>
									{user?.role === 'admin' && (
										<button 
											onClick={(e) => { e.stopPropagation(); handleDeleteIssue(issue.id) }}
											style={{ background: 'none', border: 'none', color: 'var(--text4)', cursor: 'pointer', padding: '4px' }}
											onMouseOver={e => e.currentTarget.style.color = '#ef4444'}
											onMouseOut={e => e.currentTarget.style.color = 'var(--text4)'}
										>
											<Trash2 size={13} />
										</button>
									)}
								</div>
							</div>
						))}
					</div>
				)}
			</div>

			{/* Right panel - Details View */}
			{selected && (
				<div style={{
					width: '380px', flexShrink: 0,
					background: 'rgba(255,255,255,0.015)', border: '1px solid rgba(255,255,255,0.06)',
					borderRadius: '12px', padding: '20px',
					display: 'flex', flexDirection: 'column', gap: '16px',
					overflowY: 'auto', height: '100%',
				}}>
					<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
						<h2 style={{ fontSize: '16px', fontWeight: 600, letterSpacing: '-0.165px', flex: 1, paddingRight: '8px', margin: 0 }}>
							{selected.title}
						</h2>
						<button onClick={() => setSelected(null)} style={{
							background: 'none', border: 'none', color: 'var(--text4)', cursor: 'pointer', display: 'flex', padding: '4px'
						}}>
							<X size={16} />
						</button>
					</div>

					<div style={{ display: 'flex', gap: '6px' }}>
						<span style={{
							padding: '3px 10px',
							background: `${STATUS_COLOR[selected.status]}12`, color: STATUS_COLOR[selected.status],
							borderRadius: '9999px', fontSize: '11px', fontWeight: 600, textTransform: 'uppercase'
						}}>
							{STATUS_LABEL[selected.status]}
						</span>
					</div>

					{/* Image Preview (MinIO URL) */}
					{selected.photo_url && (
						<div style={{ borderRadius: '8px', overflow: 'hidden', border: '1px solid rgba(255,255,255,0.06)', background: '#000' }}>
							<a href={selected.photo_url} target="_blank" rel="noopener noreferrer" title="Click to view full image">
								<img 
									src={selected.photo_url} 
									alt="Attached context" 
									style={{ width: '100%', maxHeight: '200px', objectFit: 'contain', display: 'block', cursor: 'zoom-in' }} 
								/>
							</a>
							<div style={{ padding: '6px 12px', background: 'rgba(0,0,0,0.4)', fontSize: '11px', color: 'var(--text4)', display: 'flex', alignItems: 'center', gap: '4px' }}>
								<Camera size={12} /> Click image to enlarge (MinIO Storage)
							</div>
						</div>
					)}

					<p style={{ fontSize: '13px', color: 'var(--text3)', lineHeight: 1.6, whiteSpace: 'pre-wrap', margin: 0 }}>
						{selected.description}
					</p>

					<div style={{ borderTop: '1px solid rgba(255,255,255,0.04)', paddingTop: '12px', display: 'flex', flexDirection: 'column', gap: '6px', fontSize: '12px', color: 'var(--text4)' }}>
						<div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
							<MapPin size={12} />
							<span>Room: <strong style={{ color: 'var(--text3)' }}>{selected.room_number}</strong></span>
						</div>
						<div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
							<Calendar size={12} />
							<span>Created: <strong style={{ color: 'var(--text3)' }}>{new Date(selected.created_at).toLocaleString()}</strong></span>
						</div>
					</div>

					{/* Staff Actions Section */}
					{isStaff && (
						<div style={{ borderTop: '1px solid rgba(255,255,255,0.04)', paddingTop: '14px', display: 'flex', flexDirection: 'column', gap: '12px' }}>
							<h4 style={{ fontSize: '11px', fontWeight: 600, color: 'var(--text4)', letterSpacing: '0.05em', textTransform: 'uppercase', margin: 0 }}>
								Staff Actions
							</h4>
							
							{/* Status Select */}
							<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
								<label style={{ fontSize: '12px', color: 'var(--text3)' }}>Update Status</label>
								<select
									value={selected.status}
									onChange={e => handleStatusChange(e.target.value)}
									style={{ ...inputStyle, padding: '8px 10px', fontSize: '13px' }}
								>
									<option value="open">Open</option>
									<option value="in_progress">In Progress</option>
									<option value="resolved">Resolved</option>
									<option value="closed">Closed</option>
								</select>
							</div>

							{/* Assign Worker Select */}
							<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
								<label style={{ fontSize: '12px', color: 'var(--text3)' }}>Assign Maintenance Worker</label>
								<select
									value={selected.worker_id || ''}
									onChange={e => handleAssignWorker(e.target.value)}
									style={{ ...inputStyle, padding: '8px 10px', fontSize: '13px' }}
								>
									<option value="">Unassigned</option>
									{workers.map(w => (
										<option key={w.id} value={w.id}>{w.name} ({w.specialty})</option>
									))}
								</select>
							</div>
						</div>
					)}

					{/* Comments Section */}
					<div style={{ borderTop: '1px solid rgba(255,255,255,0.04)', paddingTop: '14px', flex: 1, display: 'flex', flexDirection: 'column', gap: '10px' }}>
						<h4 style={{ fontSize: '11px', fontWeight: 600, color: 'var(--text4)', letterSpacing: '0.05em', textTransform: 'uppercase', margin: 0, display: 'flex', alignItems: 'center', gap: '6px' }}>
							<MessageSquare size={12} /> Comments {comments.length > 0 && `(${comments.length})`}
						</h4>
						
						<div style={{ display: 'flex', flexDirection: 'column', gap: '8px', overflowY: 'auto', flex: 1, maxHeight: '180px' }}>
							{comments.length === 0 ? (
								<div style={{ fontSize: '12px', color: 'var(--text4)', fontStyle: 'italic' }}>No comments yet</div>
							) : comments.map((c: any) => (
								<div key={c.id} style={{
									background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(255,255,255,0.04)', borderRadius: '8px', padding: '8px 10px',
								}}>
									<div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '10px', color: 'var(--text4)', marginBottom: '3px' }}>
										<span>{c.user_id === user?.id ? 'You' : 'Staff'}</span>
										<span>{new Date(c.created_at).toLocaleDateString()}</span>
									</div>
									<div style={{ fontSize: '12.5px', color: 'var(--text2)' }}>{c.text}</div>
								</div>
							))}
						</div>
						
						<div style={{ display: 'flex', gap: '6px', flexShrink: 0 }}>
							<input
								style={{ ...inputStyle, flex: 1, fontSize: '13px', padding: '8px 12px' }}
								placeholder="Write a message..."
								value={comment}
								onChange={e => setComment(e.target.value)}
								onKeyDown={e => e.key === 'Enter' && handleComment()}
							/>
							<button onClick={handleComment} style={{
								padding: '8px 14px', background: 'var(--accent)',
								border: 'none', borderRadius: '8px',
								color: '#fff', fontSize: '12px', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center'
							}}>
								<Send size={12} />
							</button>
						</div>
					</div>
				</div>
			)}
		</div>
	)
}
