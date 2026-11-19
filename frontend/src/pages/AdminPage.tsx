import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import { issuesAPI, roomsAPI, residentsAPI, authAPI } from '../api'
import type { Issue, Worker } from '../api'
import { 
	Settings, 
	Users, 
	Home, 
	Wrench, 
	ShieldAlert, 
	Plus, 
	UserMinus
} from 'lucide-react'

const STATUS_COLOR: Record<string, string> = {
	open: '#7170ff', in_progress: '#f59e0b', resolved: '#10b981', closed: '#62666d',
}

const inputStyle = {
	padding: '8px 12px',
	background: 'rgba(255,255,255,0.02)',
	border: '1px solid rgba(255,255,255,0.06)',
	borderRadius: '8px', color: 'var(--text2)',
	fontSize: '13px', outline: 'none',
}

export default function AdminPage() {
	const { user } = useAuth()
	const navigate = useNavigate()
	const [activeTab, setActiveTab] = useState<'issues' | 'residents' | 'rooms'>('issues')
	const [expandedRoom, setExpandedRoom] = useState<string | null>(null)
	
	// Data states
	const [issues, setIssues] = useState<Issue[]>([])
	const [residents, setResidents] = useState<any[]>([])
	const [rooms, setRooms] = useState<any[]>([])
	const [workers, setWorkers] = useState<Worker[]>([])
	const [loading, setLoading] = useState(true)

	// Form states for creating room
	const [newRoom, setNewRoom] = useState({ room_number: '', floor: 1, capacity: 4 })
	const [showAddRoom, setShowAddRoom] = useState(false)

	// Form states for creating user
	const [showCreateUser, setShowCreateUser] = useState(false)
	const [newUserForm, setNewUserForm] = useState({
		fullName: '',
		email: '',
		password: '',
		role: 'student'
	})

	const isAdmin = user?.role === 'admin'
	const isStaff = user?.role === 'manager' || user?.role === 'admin'

	// If a standard student tries to access, deny access immediately
	if (!isStaff) {
		return (
			<div style={{
				maxWidth: '500px', margin: '80px auto', textAlign: 'center',
				padding: '40px', background: 'rgba(255,255,255,0.01)',
				border: '1px solid rgba(239, 68, 68, 0.2)', borderRadius: '12px'
			}}>
				<ShieldAlert size={48} style={{ color: '#ef4444', marginBottom: '16px' }} />
				<h2 style={{ fontSize: '18px', fontWeight: 600, color: 'var(--text)', marginBottom: '8px' }}>Access Denied</h2>
				<p style={{ fontSize: '14px', color: 'var(--text3)', lineHeight: 1.5, marginBottom: '24px' }}>
					You do not have the required staff privileges (Manager or Admin) to view the DormOS Administration Panel.
				</p>
				<button 
					onClick={() => navigate('/dashboard')}
					style={{
						padding: '8px 16px', background: 'var(--accent)', border: 'none',
						borderRadius: '8px', color: '#fff', fontSize: '13px', fontWeight: 510, cursor: 'pointer'
					}}
				>
					Return to Dashboard
				</button>
			</div>
		)
	}

	const loadAllData = async () => {
		setLoading(true)
		try {
			// Load all issues
			const issRes = await issuesAPI.listAll()
			setIssues(issRes.data.issues || [])
			
			// Load workers for issues
			const workersRes = await issuesAPI.listWorkers()
			setWorkers(workersRes.data.workers || [])

			// Load residents & rooms if user is admin
			if (isAdmin) {
				const resRes = await residentsAPI.list()
				setResidents(resRes.data.residents || [])
				
				const roomsRes = await roomsAPI.list()
				setRooms(roomsRes.data.rooms || [])
			}
		} catch (err) {
			console.error('Failed to load admin panel data:', err)
		} finally {
			setLoading(false)
		}
	}

	useEffect(() => {
		loadAllData()
	}, [user])

	const handleStatusChange = async (issueId: string, newStatus: string) => {
		try {
			await issuesAPI.updateStatus(issueId, newStatus)
			loadAllData()
		} catch (err) {
			console.error('Failed to update status:', err)
		}
	}

	const handleAssignWorker = async (issueId: string, workerId: string) => {
		try {
			await issuesAPI.assignWorker(issueId, workerId)
			loadAllData()
			alert('Worker assigned successfully!')
		} catch (err) {
			console.error('Failed to assign worker:', err)
		}
	}

	// Admin-Only actions
	const handleRoleChange = async (userId: string, role: string) => {
		if (!isAdmin) return
		try {
			await residentsAPI.updateRole(userId, role)
			loadAllData()
			alert(`Role updated to ${role.toUpperCase()} successfully!`)
		} catch (err) {
			console.error('Failed to update role:', err)
		}
	}

	const handleEvictResident = async (userId: string) => {
		if (!isAdmin) return
		if (!window.confirm('Are you sure you want to evict this resident from their room?')) return
		try {
			await residentsAPI.remove(userId)
			loadAllData()
			alert('Resident successfully evicted!')
		} catch (err) {
			console.error('Failed to remove resident:', err)
		}
	}

	const handleCreateRoom = async (e: React.FormEvent) => {
		e.preventDefault()
		if (!isAdmin) return
		try {
			await roomsAPI.create(newRoom)
			setNewRoom({ room_number: '', floor: 1, capacity: 4 })
			setShowAddRoom(false)
			loadAllData()
			alert('New Room successfully created!')
		} catch (err) {
			console.error('Failed to create room:', err)
		}
	}

	const handleCreateUser = async (e: React.FormEvent) => {
		e.preventDefault()
		if (!isAdmin) return
		try {
			const res = await authAPI.register({
				full_name: newUserForm.fullName,
				email: newUserForm.email,
				password: newUserForm.password
			})
			const newUserId = res.data.user_id

			if (newUserForm.role !== 'student') {
				await residentsAPI.updateRole(newUserId, newUserForm.role)
			}

			setNewUserForm({
				fullName: '',
				email: '',
				password: '',
				role: 'student'
			})
			setShowCreateUser(false)
			loadAllData()
			alert('New User/Staff successfully registered!')
		} catch (err) {
			console.error('Failed to create user:', err)
			alert('Failed to register user. Make sure email is unique and password is >= 8 chars.')
		}
	}

	return (
		<div style={{ maxWidth: '1000px' }}>
			{/* Header */}
			<div style={{ marginBottom: '28px', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
				<div>
					<h1 style={{ fontSize: '22px', fontWeight: 600, letterSpacing: '-0.24px', marginBottom: '4px', display: 'flex', alignItems: 'center', gap: '8px' }}>
						<Settings size={22} className="text-accent" />
						Admin Panel
					</h1>
					<p style={{ fontSize: '13px', color: 'var(--text4)' }}>
						{isAdmin ? 'System Administrator control center' : 'Dormitory Floor Manager desk'}
					</p>
				</div>
				<button 
					onClick={loadAllData}
					style={{
						padding: '6px 12px', background: 'transparent',
						border: '1px solid rgba(255,255,255,0.08)', borderRadius: '6px',
						color: 'var(--text3)', fontSize: '12px', cursor: 'pointer'
					}}
				>
					Refresh Data
				</button>
			</div>

			{/* Navigation Tabs */}
			<div style={{ display: 'flex', borderBottom: '1px solid rgba(255,255,255,0.06)', marginBottom: '20px', gap: '16px' }}>
				<button 
					onClick={() => setActiveTab('issues')}
					style={{
						padding: '8px 12px', background: 'none', border: 'none',
						borderBottom: activeTab === 'issues' ? '2px solid var(--accent)' : '2px solid transparent',
						color: activeTab === 'issues' ? 'var(--text)' : 'var(--text4)',
						fontWeight: activeTab === 'issues' ? 600 : 400,
						cursor: 'pointer', fontSize: '14px', display: 'flex', alignItems: 'center', gap: '8px'
					}}
				>
					<Wrench size={16} /> All Issues ({issues.length})
				</button>

				{isAdmin && (
					<>
						<button 
							onClick={() => setActiveTab('residents')}
							style={{
								padding: '8px 12px', background: 'none', border: 'none',
								borderBottom: activeTab === 'residents' ? '2px solid var(--accent)' : '2px solid transparent',
								color: activeTab === 'residents' ? 'var(--text)' : 'var(--text4)',
								fontWeight: activeTab === 'residents' ? 600 : 400,
								cursor: 'pointer', fontSize: '14px', display: 'flex', alignItems: 'center', gap: '8px'
							}}
						>
							<Users size={16} /> Residents ({residents.length})
						</button>

						<button 
							onClick={() => setActiveTab('rooms')}
							style={{
								padding: '8px 12px', background: 'none', border: 'none',
								borderBottom: activeTab === 'rooms' ? '2px solid var(--accent)' : '2px solid transparent',
								color: activeTab === 'rooms' ? 'var(--text)' : 'var(--text4)',
								fontWeight: activeTab === 'rooms' ? 600 : 400,
								cursor: 'pointer', fontSize: '14px', display: 'flex', alignItems: 'center', gap: '8px'
							}}
						>
							<Home size={16} /> Rooms ({rooms.length})
						</button>
					</>
				)}
			</div>

			{loading ? (
				<div style={{ color: 'var(--text4)', fontSize: '13px', textAlign: 'center', padding: '48px' }}>Loading registry information...</div>
			) : (
				<>
					{/* Tab 1: Issues */}
					{activeTab === 'issues' && (
						<div style={{
							background: 'rgba(255,255,255,0.01)', border: '1px solid rgba(255,255,255,0.06)',
							borderRadius: '12px', overflow: 'hidden',
						}}>
							<div style={{
								display: 'grid', gridTemplateColumns: '1fr 80px 100px 150px 130px',
								padding: '12px 18px', background: 'rgba(255,255,255,0.01)',
								borderBottom: '1px solid rgba(255,255,255,0.05)',
							}}>
								{['Issue / Details', 'Room', 'Status', 'Assigned Worker', 'Actions'].map(h => (
									<div key={h} style={{ fontSize: '11px', fontWeight: 600, color: 'var(--text4)', letterSpacing: '0.05em', textTransform: 'uppercase' }}>
										{h}
									</div>
								))}
							</div>

							{issues.length === 0 ? (
								<div style={{ padding: '40px', textAlign: 'center', color: 'var(--text4)', fontSize: '13px' }}>
									No reported issues in the system.
								</div>
							) : issues.map((issue, i) => (
								<div
									key={issue.id}
									style={{
										display: 'grid', gridTemplateColumns: '1fr 80px 100px 150px 130px',
										padding: '14px 18px', alignItems: 'center',
										borderBottom: i < issues.length - 1 ? '1px solid rgba(255,255,255,0.04)' : 'none',
									}}
								>
									<div style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', paddingRight: '12px' }}>
										<div style={{ fontSize: '13.5px', fontWeight: 600, color: 'var(--text)', marginBottom: '4px' }}>
											{issue.title}
										</div>
										<div style={{ fontSize: '11.5px', color: 'var(--text4)', overflow: 'hidden', textOverflow: 'ellipsis' }}>
											{issue.description}
										</div>
									</div>
									<div style={{ fontSize: '13px', color: 'var(--text2)', fontWeight: 500 }}>{issue.room_number}</div>
									<div>
										<span style={{
											display: 'inline-block', padding: '3px 10px',
											background: `${STATUS_COLOR[issue.status]}12`,
											color: STATUS_COLOR[issue.status],
											borderRadius: '9999px', fontSize: '11px', fontWeight: 600,
											textTransform: 'uppercase'
										}}>
											{issue.status.replace('_', ' ')}
										</span>
									</div>
									
									{/* Worker Assignment (Student-Scoped 12 Endpoint check!) */}
									<div>
										<select
											value={issue.worker_id || ''}
											onChange={e => handleAssignWorker(issue.id, e.target.value)}
											style={{ ...inputStyle, width: '90%', fontSize: '12px', padding: '4px 6px' }}
										>
											<option value="">Unassigned</option>
											{workers.map(w => (
												<option key={w.id} value={w.id}>{w.name}</option>
											))}
										</select>
									</div>

									{/* Status Change */}
									<div>
										<select
											value={issue.status}
											onChange={e => handleStatusChange(issue.id, e.target.value)}
											style={{
												...inputStyle, fontSize: '12px', padding: '5px 8px', width: '100%',
												background: 'rgba(255,255,255,0.04)'
											}}
										>
											<option value="open">Open</option>
											<option value="in_progress">In Progress</option>
											<option value="resolved">Resolved</option>
											<option value="closed">Closed</option>
										</select>
									</div>
								</div>
							))}
						</div>
					)}

					{/* Tab 2: Residents (Admin Only) */}
					{activeTab === 'residents' && isAdmin && (
						<div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
							
							{/* Create User Form */}
							<div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
								<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
									<h2 style={{ fontSize: '15px', fontWeight: 600, color: 'var(--text2)' }}>Dormitory Residents Registry</h2>
									<button 
										onClick={() => setShowCreateUser(!showCreateUser)}
										style={{
											padding: '6px 12px', background: 'var(--accent)', border: 'none',
											borderRadius: '6px', color: '#fff', fontSize: '12px', fontWeight: 510,
											cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px'
										}}
									>
										<Plus size={14} /> Add User/Staff
									</button>
								</div>

								{showCreateUser && (
									<form onSubmit={handleCreateUser} style={{
										display: 'flex', gap: '12px', background: 'rgba(255,255,255,0.01)',
										border: '1px solid rgba(255,255,255,0.06)', borderRadius: '8px', padding: '16px',
										alignItems: 'center', flexWrap: 'wrap'
									}}>
										<input 
											style={{ ...inputStyle, flex: 1, minWidth: '150px' }} placeholder="Full Name" required
											value={newUserForm.fullName} onChange={e => setNewUserForm({ ...newUserForm, fullName: e.target.value })}
										/>
										<input 
											type="email" style={{ ...inputStyle, flex: 1, minWidth: '180px' }} placeholder="Email Address" required
											value={newUserForm.email} onChange={e => setNewUserForm({ ...newUserForm, email: e.target.value })}
										/>
										<input 
											type="password" style={{ ...inputStyle, flex: 1, minWidth: '150px' }} placeholder="Password (min 8 chars)" required
											value={newUserForm.password} onChange={e => setNewUserForm({ ...newUserForm, password: e.target.value })}
										/>
										<select 
											style={{ ...inputStyle, width: '150px' }}
											value={newUserForm.role} onChange={e => setNewUserForm({ ...newUserForm, role: e.target.value })}
										>
											<option value="student">Student</option>
											<option value="manager">Manager</option>
											<option value="admin">System Admin</option>
										</select>
										<button 
											type="submit"
											style={{
												padding: '8px 16px', background: 'var(--success)', border: 'none',
												borderRadius: '6px', color: '#fff', fontSize: '12px', fontWeight: 510,
												cursor: 'pointer'
											}}
										>
											Create Account
										</button>
									</form>
								)}
							</div>

							<div style={{
								background: 'rgba(255,255,255,0.01)', border: '1px solid rgba(255,255,255,0.06)',
								borderRadius: '12px', overflow: 'hidden',
							}}>
								<div style={{
									display: 'grid', gridTemplateColumns: '1fr 180px 140px 100px',
									padding: '12px 18px', background: 'rgba(255,255,255,0.01)',
									borderBottom: '1px solid rgba(255,255,255,0.05)',
								}}>
									{['Resident UUID & Info', 'Room Assignment', 'System Role', 'Actions'].map(h => (
										<div key={h} style={{ fontSize: '11px', fontWeight: 600, color: 'var(--text4)', letterSpacing: '0.05em', textTransform: 'uppercase' }}>
											{h}
										</div>
									))}
								</div>

								{residents.length === 0 ? (
									<div style={{ padding: '40px', textAlign: 'center', color: 'var(--text4)', fontSize: '13px' }}>
										No registered residents or users found.
									</div>
								) : residents.map((res, i) => {
									const parts = (res.user_id || '').split(':')
									const uuid = parts[0] || res.user_id || ''
									const fullName = parts[1] || 'Registered Student'
									const email = parts[2] || ''
									const hasRoom = res.room_number && res.room_number !== 'Unassigned'

									return (
										<div
											key={uuid + '-' + i}
											style={{
												display: 'grid', gridTemplateColumns: '1fr 180px 140px 100px',
												padding: '14px 18px', alignItems: 'center',
												borderBottom: i < residents.length - 1 ? '1px solid rgba(255,255,255,0.04)' : 'none',
											}}
										>
											<div style={{ display: 'flex', flexDirection: 'column', gap: '2px', minWidth: 0 }}>
												<div style={{ fontSize: '13px', fontWeight: 510, color: 'var(--text)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
													{fullName}
												</div>
												{email && (
													<div style={{ fontSize: '11px', color: 'var(--text4)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
														{email}
													</div>
												)}
												<div style={{ fontFamily: 'monospace', fontSize: '10px', color: 'rgba(255,255,255,0.2)' }}>
													ID: {uuid}
												</div>
											</div>

											<div style={{ fontSize: '13px', color: 'var(--text)' }}>
												{hasRoom ? (
													<span style={{ fontWeight: 500, color: 'var(--success)' }}>Rm {res.room_number}</span>
												) : (
													<select
														defaultValue=""
														onChange={async (e) => {
															const roomId = e.target.value
															if (!roomId) return
															try {
																await residentsAPI.assign(uuid, roomId)
																loadAllData()
																alert('Resident successfully assigned to room!')
															} catch (err) {
																console.error(err)
																alert('Failed to assign room')
															}
														}}
														style={{
															...inputStyle, padding: '4px 6px', fontSize: '12px', width: '150px',
															background: 'rgba(255,255,255,0.04)', color: 'var(--text2)',
															border: '1px solid rgba(255,255,255,0.1)'
														}}
													>
														<option value="">Assign Room...</option>
														{rooms.map(rm => (
															<option key={rm.room_id} value={rm.room_id}>Room {rm.room_number} ({rm.occupied || 0}/{rm.capacity})</option>
														))}
													</select>
												)}
											</div>
											
											{/* Update Resident Role */}
											<div>
												<select
													value={res.role || 'student'}
													onChange={e => handleRoleChange(uuid, e.target.value)}
													style={{
														...inputStyle, fontSize: '12px', padding: '4px 8px',
														background: 'rgba(255,255,255,0.04)', color: 'var(--text2)',
														border: '1px solid rgba(255,255,255,0.1)'
													}}
												>
													<option value="student">Student</option>
													<option value="manager">Manager</option>
													<option value="admin">System Admin</option>
												</select>
											</div>

											{/* Remove Resident */}
											<div>
												{hasRoom ? (
													<button
														onClick={() => handleEvictResident(uuid)}
														style={{
															padding: '6px 10px', background: 'rgba(239, 68, 68, 0.1)',
															border: '1px solid rgba(239, 68, 68, 0.2)', borderRadius: '6px',
															color: '#ef4444', fontSize: '11px', cursor: 'pointer',
															display: 'flex', alignItems: 'center', gap: '4px'
														}}
													>
														<UserMinus size={12} /> Evict
													</button>
												) : (
													<span style={{ fontSize: '11px', color: 'var(--text4)', fontStyle: 'italic' }}>Unassigned</span>
												)}
											</div>
										</div>
									)
								})}
							</div>
						</div>
					)}

					{/* Tab 3: Rooms (Admin Only) */}
					{activeTab === 'rooms' && isAdmin && (
						<div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
							
							{/* Create Room Form (Student-Scoped 12 Endpoint check!) */}
							<div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
								<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
									<h2 style={{ fontSize: '15px', fontWeight: 600, color: 'var(--text2)' }}>Dormitory Rooms Registry</h2>
									<button 
										onClick={() => setShowAddRoom(!showAddRoom)}
										style={{
											padding: '6px 12px', background: 'var(--accent)', border: 'none',
											borderRadius: '6px', color: '#fff', fontSize: '12px', fontWeight: 510,
											cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px'
										}}
									>
										<Plus size={14} /> Add Room
									</button>
								</div>

								{showAddRoom && (
									<form onSubmit={handleCreateRoom} style={{
										display: 'flex', gap: '12px', background: 'rgba(255,255,255,0.01)',
										border: '1px solid rgba(255,255,255,0.06)', borderRadius: '8px', padding: '16px',
										alignItems: 'center'
									}}>
										<input 
											style={inputStyle} placeholder="Room (e.g. 101)" required
											value={newRoom.room_number} onChange={e => setNewRoom({ ...newRoom, room_number: e.target.value })}
										/>
										<input 
											type="number" style={inputStyle} placeholder="Floor" required min={1}
											value={newRoom.floor} onChange={e => setNewRoom({ ...newRoom, floor: parseInt(e.target.value) || 1 })}
										/>
										<input 
											type="number" style={inputStyle} placeholder="Max Capacity" required min={1}
											value={newRoom.capacity} onChange={e => setNewRoom({ ...newRoom, capacity: parseInt(e.target.value) || 4 })}
										/>
										<button 
											type="submit"
											style={{
												padding: '8px 16px', background: 'var(--accent)', border: 'none',
												borderRadius: '6px', color: '#fff', fontSize: '12px', fontWeight: 510, cursor: 'pointer'
											}}
										>
											Create Room
										</button>
										<button 
											type="button" onClick={() => setShowAddRoom(false)}
											style={{
												padding: '8px 12px', background: 'transparent', border: '1px solid rgba(255,255,255,0.08)',
												borderRadius: '6px', color: 'var(--text4)', fontSize: '12px', cursor: 'pointer'
											}}
										>
											Cancel
										</button>
									</form>
								)}
							</div>

							<div style={{
								background: 'rgba(255,255,255,0.01)', border: '1px solid rgba(255,255,255,0.06)',
								borderRadius: '12px', overflow: 'hidden',
							}}>
								<div style={{
									display: 'grid', gridTemplateColumns: '1fr 120px 120px 120px',
									padding: '12px 18px', background: 'rgba(255,255,255,0.01)',
									borderBottom: '1px solid rgba(255,255,255,0.05)',
								}}>
									{['Room Number', 'Floor level', 'Max Capacity', 'Current Occupants'].map(h => (
										<div key={h} style={{ fontSize: '11px', fontWeight: 600, color: 'var(--text4)', letterSpacing: '0.05em', textTransform: 'uppercase' }}>
											{h}
										</div>
									))}
								</div>

								{rooms.length === 0 ? (
									<div style={{ padding: '40px', textAlign: 'center', color: 'var(--text4)', fontSize: '13px' }}>
										No rooms registered in the system.
									</div>
								) : rooms.map((rm, i) => (
									<div key={rm.room_id} style={{ borderBottom: i < rooms.length - 1 ? '1px solid rgba(255,255,255,0.04)' : 'none' }}>
										<div
											onClick={() => setExpandedRoom(expandedRoom === rm.room_id ? null : rm.room_id)}
											style={{
												display: 'grid', gridTemplateColumns: '1fr 120px 120px 120px',
												padding: '14px 18px', alignItems: 'center',
												cursor: 'pointer', transition: 'background 0.2s'
											}}
											onMouseOver={e => e.currentTarget.style.background = 'rgba(255,255,255,0.015)'}
											onMouseOut={e => e.currentTarget.style.background = 'transparent'}
										>
											<div style={{ fontSize: '14px', fontWeight: 600, color: 'var(--text)' }}>
												Room {rm.room_number}
											</div>
											<div style={{ fontSize: '13px', color: 'var(--text3)' }}>
												Floor {rm.floor}
											</div>
											<div style={{ fontSize: '13px', color: 'var(--text3)' }}>
												{rm.capacity} beds
											</div>
											<div style={{ fontSize: '13px', color: rm.occupied >= rm.capacity ? '#ef4444' : '#10b981', fontWeight: 600 }}>
												{rm.occupied || 0} / {rm.capacity} beds filled
											</div>
										</div>

										{expandedRoom === rm.room_id && (
											<div style={{
												padding: '12px 18px 18px 18px',
												background: 'rgba(255,255,255,0.005)',
												borderTop: '1px solid rgba(255,255,255,0.03)',
												display: 'flex',
												flexDirection: 'column',
												gap: '8px'
											}}>
												<div style={{ fontSize: '10px', fontWeight: 600, color: 'var(--text4)', textTransform: 'uppercase', marginBottom: '4px', letterSpacing: '0.05em' }}>
													Assigned Room Residents
												</div>
												{!rm.residents || rm.residents.length === 0 ? (
													<div style={{ fontSize: '12px', color: 'var(--text4)', fontStyle: 'italic', paddingLeft: '4px' }}>
														No occupants assigned to this room yet.
													</div>
												) : (
													<div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
														{rm.residents.map((res: any) => (
															<div key={res.user_id} style={{
																display: 'flex',
																justifyContent: 'space-between',
																alignItems: 'center',
																background: 'rgba(255,255,255,0.01)',
																border: '1px solid rgba(255,255,255,0.04)',
																borderRadius: '8px',
																padding: '8px 12px'
															}}>
																<div>
																	<div style={{ fontSize: '12.5px', fontWeight: 510, color: 'var(--text2)' }}>
																		{res.full_name} <span style={{ fontSize: '11px', color: 'var(--text4)', marginLeft: '6px' }}>({res.role})</span>
																	</div>
																	<div style={{ fontSize: '10px', fontFamily: 'monospace', color: 'rgba(255,255,255,0.2)', marginTop: '2px' }}>
																		ID: {res.user_id}
																	</div>
																</div>
																<button
																	onClick={async (e) => {
																		e.stopPropagation()
																		if (!window.confirm(`Are you sure you want to evict ${res.full_name} from Room ${rm.room_number}?`)) return
																		try {
																			await residentsAPI.remove(res.user_id)
																			loadAllData()
																			alert('Resident successfully evicted!')
																		} catch (err) {
																			console.error(err)
																			alert('Failed to evict resident')
																		}
																	}}
																	style={{
																		padding: '4px 8px', background: 'rgba(239, 68, 68, 0.08)',
																		border: '1px solid rgba(239, 68, 68, 0.15)', borderRadius: '6px',
																		color: '#ef4444', fontSize: '11px', cursor: 'pointer',
																		transition: 'background 0.2s'
																	}}
																	onMouseOver={e => e.currentTarget.style.background = 'rgba(239, 68, 68, 0.15)'}
																	onMouseOut={e => e.currentTarget.style.background = 'rgba(239, 68, 68, 0.08)'}
																>
																	Evict
																</button>
															</div>
														))}
													</div>
												)}
											</div>
										)}
									</div>
								))}
							</div>
						</div>
					)}
				</>
			)}
		</div>
	)
}
