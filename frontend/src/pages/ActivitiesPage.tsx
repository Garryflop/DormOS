import { useEffect, useState } from 'react'
import { useAuth } from '../AuthContext'
import { eventsAPI } from '../api'
import { Star, Target, Trophy, Calendar, MapPin, Plus, Trash2, X, Check, Users } from 'lucide-react'

interface Event {
	event_id: string
	title: string
	description: string
	location: string
	start_time: number
	end_time: number
	max_participants: number
	registered_count: number
	created_at: number
}

const inputStyle = {
	width: '100%', padding: '10px 14px',
	background: 'rgba(255,255,255,0.02)',
	border: '1px solid rgba(255,255,255,0.06)',
	borderRadius: '8px', color: 'var(--text2)',
	fontSize: '14px', outline: 'none',
	transition: 'border-color 0.15s',
}

export default function ActivitiesPage() {
	const { user } = useAuth()
	const [events, setEvents] = useState<Event[]>([])
	const [loading, setLoading] = useState(true)
	const [showForm, setShowForm] = useState(false)
	const [submitting, setSubmitting] = useState(false)

	// Form states
	const [form, setForm] = useState({
		title: '',
		description: '',
		location: '',
		max_participants: 50,
		start_date: '',
		start_time: '18:00',
	})

	const isAdmin = user?.role === 'admin'

	const loadEvents = async () => {
		setLoading(true)
		try {
			const res = await eventsAPI.list()
			setEvents(res.data.events || [])
		} catch (err) {
			console.error('Failed to load events:', err)
		} finally {
			setLoading(false)
		}
	}

	useEffect(() => {
		loadEvents()
	}, [])

	const handleCreateEvent = async (e: React.FormEvent) => {
		e.preventDefault()
		setSubmitting(true)
		try {
			// Convert start_date and start_time to unix timestamp
			const dateTimeStr = `${form.start_date}T${form.start_time}`
			const timestamp = Math.floor(new Date(dateTimeStr).getTime() / 1000)

			await eventsAPI.create({
				title: form.title,
				description: form.description,
				location: form.location,
				start_time: timestamp,
				end_time: timestamp + 7200, // 2 hours duration default
				max_participants: Number(form.max_participants),
			})

			setShowForm(false)
			setForm({ title: '', description: '', location: '', max_participants: 50, start_date: '', start_time: '18:00' })
			await loadEvents()
		} catch (err) {
			console.error('Failed to create event:', err)
			alert('Failed to create event. Make sure details are valid.')
		} finally {
			setSubmitting(false)
		}
	}

	const handleDeleteEvent = async (eventId: string) => {
		if (!window.confirm('Are you sure you want to delete this event?')) return
		try {
			await eventsAPI.delete(eventId)
			await loadEvents()
		} catch (err) {
			console.error('Failed to delete event:', err)
		}
	}

	const handleRegister = async (eventId: string) => {
		try {
			await eventsAPI.register(eventId)
			await loadEvents()
			alert('Successfully registered for the event!')
		} catch (err) {
			console.error(err)
			alert('Failed to register. Maybe event is full or already registered.')
		}
	}

	const handleCancel = async (eventId: string) => {
		try {
			await eventsAPI.cancel(eventId)
			await loadEvents()
			alert('Successfully cancelled registration.')
		} catch (err) {
			console.error(err)
			alert('Failed to cancel registration.')
		}
	}

	// Helper to extract clean description and registered status
	const parseEventInfo = (evt: Event) => {
		const desc = evt.description || ''
		const match = desc.match(/\[participants:(.*)\]/)
		const participantList = match && match[1] ? match[1].split(',') : []
		const cleanDesc = desc.replace(/\n\n\[participants:.*\]/, '')
		const isRegistered = user ? participantList.includes(user.id) : false
		return { cleanDesc, isRegistered }
	}

	// Dynamic stats
	const attendedEventsCount = events.filter(e => {
		const { isRegistered } = parseEventInfo(e)
		return isRegistered
	}).length

	const totalPoints = attendedEventsCount * 15 // 15 points per event!

	return (
		<div style={{ width: '100%', display: 'flex', flexDirection: 'column', gap: '24px' }}>
			{/* Header */}
			<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
				<div>
					<h1 style={{ fontSize: '22px', fontWeight: 600, letterSpacing: '-0.24px', marginBottom: '4px' }}>Activities Portal</h1>
					<p style={{ fontSize: '13px', color: 'var(--text4)' }}>Register for community events, meet neighbors, and track your activity score</p>
				</div>
				{isAdmin && (
					<button
						onClick={() => setShowForm(!showForm)}
						style={{
							padding: '8px 16px', background: 'var(--accent)',
							border: 'none', borderRadius: '8px',
							color: '#fff', fontSize: '13px', fontWeight: 510, cursor: 'pointer',
							letterSpacing: '-0.13px', display: 'flex', alignItems: 'center', gap: '8px',
							transition: 'opacity 0.15s'
						}}
					>
						<Plus size={16} /> Create Event
					</button>
				)}
			</div>

			{/* Stats Widgets */}
			<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
				{[
					{ label: 'Dormitory Points', value: loading ? '—' : `${totalPoints} XP`, icon: Star, color: '#f59e0b', subtext: 'Earned from active participation' },
					{ label: 'Your Event Registrations', value: loading ? '—' : `${attendedEventsCount} Events`, icon: Target, color: '#7170ff', subtext: 'Upcoming scheduled gatherings' },
				].map(card => {
					const Icon = card.icon
					return (
						<div key={card.label} style={{
							background: 'rgba(255,255,255,0.015)', border: '1px solid rgba(255,255,255,0.06)',
							borderRadius: '12px', padding: '20px 24px',
							display: 'flex', alignItems: 'center', gap: '20px',
						}}>
							<div style={{ 
								background: `${card.color}12`, 
								padding: '12px', 
								borderRadius: '12px', 
								color: card.color,
								display: 'flex',
								alignItems: 'center',
								justifyContent: 'center'
							}}>
								<Icon size={22} />
							</div>
							<div>
								<div style={{ fontSize: '26px', fontWeight: 600, letterSpacing: '-0.5px', color: 'var(--text)', marginBottom: '2px' }}>
									{card.value}
								</div>
								<div style={{ fontSize: '12.5px', color: 'var(--text3)', fontWeight: 500, marginBottom: '2px' }}>{card.label}</div>
								<div style={{ fontSize: '11px', color: 'var(--text4)' }}>{card.subtext}</div>
							</div>
						</div>
					)
				})}
			</div>

			{/* Create Event Form Modal/Box */}
			{showForm && isAdmin && (
				<div style={{
					background: 'rgba(255,255,255,0.015)', border: '1px solid rgba(113,112,255,0.2)',
					borderRadius: '12px', padding: '24px', display: 'flex', flexDirection: 'column', gap: '20px'
				}}>
					<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
						<h3 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--text)' }}>Create a New Community Event</h3>
						<button onClick={() => setShowForm(false)} style={{ background: 'none', border: 'none', color: 'var(--text4)', cursor: 'pointer' }}>
							<X size={18} />
						</button>
					</div>

					<form onSubmit={handleCreateEvent} style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
						<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
							<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
								<label style={{ fontSize: '12px', color: 'var(--text3)' }}>Event Title</label>
								<input 
									style={inputStyle} placeholder="e.g. Table Tennis Championship" required
									value={form.title} onChange={e => setForm({ ...form, title: e.target.value })}
								/>
							</div>
							<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
								<label style={{ fontSize: '12px', color: 'var(--text3)' }}>Location</label>
								<input 
									style={inputStyle} placeholder="e.g. Room 204 / Hallway" required
									value={form.location} onChange={e => setForm({ ...form, location: e.target.value })}
								/>
							</div>
						</div>

						<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
							<label style={{ fontSize: '12px', color: 'var(--text3)' }}>Event Description</label>
							<textarea 
								style={{ ...inputStyle, minHeight: '80px', resize: 'vertical' }} 
								placeholder="What will happen at the event?..." required
								value={form.description} onChange={e => setForm({ ...form, description: e.target.value })}
							/>
						</div>

						<div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '16px' }}>
							<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
								<label style={{ fontSize: '12px', color: 'var(--text3)' }}>Date</label>
								<input 
									type="date" style={inputStyle} required
									value={form.start_date} onChange={e => setForm({ ...form, start_date: e.target.value })}
								/>
							</div>
							<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
								<label style={{ fontSize: '12px', color: 'var(--text3)' }}>Time</label>
								<input 
									type="time" style={inputStyle} required
									value={form.start_time} onChange={e => setForm({ ...form, start_time: e.target.value })}
								/>
							</div>
							<div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
								<label style={{ fontSize: '12px', color: 'var(--text3)' }}>Max Participants</label>
								<input 
									type="number" style={inputStyle} required min="2"
									value={form.max_participants} onChange={e => setForm({ ...form, max_participants: Number(e.target.value) })}
								/>
							</div>
						</div>

						<div style={{ display: 'flex', gap: '12px', marginTop: '4px' }}>
							<button type="submit" disabled={submitting} style={{
								padding: '10px 24px', background: 'var(--accent)', border: 'none', borderRadius: '8px',
								color: '#fff', fontSize: '13px', fontWeight: 510, cursor: 'pointer'
							}}>
								{submitting ? 'Creating Event...' : 'Publish Event'}
							</button>
							<button type="button" onClick={() => setShowForm(false)} style={{
								padding: '10px 16px', background: 'transparent', border: '1px solid rgba(255,255,255,0.08)', borderRadius: '8px',
								color: 'var(--text3)', fontSize: '13px', cursor: 'pointer'
							}}>
								Cancel
							</button>
						</div>
					</form>
				</div>
			)}

			{/* Events Listing */}
			<div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
				<h2 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--text2)', letterSpacing: '-0.165px' }}>Upcoming Dormitory Activities</h2>

				{loading ? (
					<div style={{ color: 'var(--text4)', fontSize: '13px', padding: '40px 0', textAlign: 'center' }}>Loading activities registry...</div>
				) : events.length === 0 ? (
					<div style={{
						background: 'rgba(255,255,255,0.01)', border: '1px solid rgba(255,255,255,0.06)',
						borderRadius: '12px', padding: '56px 24px', textAlign: 'center',
					}}>
						<div style={{ display: 'inline-flex', padding: '16px', background: 'rgba(255,255,255,0.02)', borderRadius: '50%', color: 'var(--text4)', marginBottom: '16px' }}>
							<Trophy size={32} />
						</div>
						<div style={{ fontSize: '14px', fontWeight: 510, color: 'var(--text2)', marginBottom: '6px' }}>No active events found</div>
						<div style={{ fontSize: '13px', color: 'var(--text4)' }}>Check back later or ask floor wardens to schedule community activities.</div>
					</div>
				) : (
					<div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))', gap: '16px' }}>
						{events.map(evt => {
							const { cleanDesc, isRegistered } = parseEventInfo(evt)
							const isFull = evt.registered_count >= evt.max_participants
							const date = new Date(evt.start_time * 1000)

							return (
								<div key={evt.event_id} style={{
									background: 'rgba(255,255,255,0.015)', border: '1px solid rgba(255,255,255,0.06)',
									borderRadius: '12px', padding: '20px', display: 'flex', flexDirection: 'column', gap: '16px',
									transition: 'border-color 0.2s', position: 'relative'
								}}>
									{/* Delete Event for Admins */}
									{isAdmin && (
										<button 
											onClick={() => handleDeleteEvent(evt.event_id)}
											style={{
												position: 'absolute', top: '16px', right: '16px',
												background: 'none', border: 'none', color: 'var(--text4)', cursor: 'pointer',
												padding: '4px'
											}}
											onMouseOver={e => e.currentTarget.style.color = '#ef4444'}
											onMouseOut={e => e.currentTarget.style.color = 'var(--text4)'}
										>
											<Trash2 size={15} />
										</button>
									)}

									{/* Event Details */}
									<div>
										<h3 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--text)', letterSpacing: '-0.154px', marginBottom: '8px', paddingRight: isAdmin ? '24px' : '0' }}>
											{evt.title}
										</h3>
										<p style={{ fontSize: '13px', color: 'var(--text3)', lineHeight: 1.5, margin: 0, minHeight: '40px' }}>
											{cleanDesc}
										</p>
									</div>

									{/* Metadata */}
									<div style={{ display: 'flex', flexDirection: 'column', gap: '6px', fontSize: '12px', color: 'var(--text4)' }}>
										<div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
											<Calendar size={13} style={{ color: 'var(--accent)' }} />
											<span>{date.toLocaleDateString()} at {date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
										</div>
										<div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
											<MapPin size={13} style={{ color: 'var(--accent)' }} />
											<span>{evt.location}</span>
										</div>
										<div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
											<Users size={13} style={{ color: 'var(--accent)' }} />
											<span>Registered: <strong style={{ color: 'var(--text3)' }}>{evt.registered_count} / {evt.max_participants}</strong></span>
										</div>
									</div>

									{/* Progress bar */}
									<div style={{ width: '100%', height: '4px', background: 'rgba(255,255,255,0.06)', borderRadius: '2px', overflow: 'hidden' }}>
										<div style={{
											width: `${Math.min(100, (evt.registered_count / evt.max_participants) * 100)}%`,
											height: '100%',
											background: isRegistered ? 'var(--success)' : 'var(--accent)',
											transition: 'width 0.3s ease'
										}} />
									</div>

									{/* Actions */}
									<div style={{ display: 'flex', gap: '10px', marginTop: '4px' }}>
										{isRegistered ? (
											<button
												onClick={() => handleCancel(evt.event_id)}
												style={{
													flex: 1, padding: '8px 12px', background: 'rgba(16, 185, 129, 0.08)',
													border: '1px solid rgba(16, 185, 129, 0.2)', borderRadius: '8px',
													color: '#10b981', fontSize: '12.5px', fontWeight: 510, cursor: 'pointer',
													display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '6px'
												}}
											>
												<Check size={14} /> Registered (Cancel)
											</button>
										) : (
											<button
												onClick={() => handleRegister(evt.event_id)}
												disabled={isFull}
												style={{
													flex: 1, padding: '8px 12px',
													background: isFull ? 'rgba(255,255,255,0.02)' : 'var(--accent)',
													border: isFull ? '1px solid rgba(255,255,255,0.04)' : 'none',
													borderRadius: '8px',
													color: isFull ? 'var(--text4)' : '#fff',
													fontSize: '12.5px', fontWeight: 510,
													cursor: isFull ? 'not-allowed' : 'pointer',
													transition: 'opacity 0.15s'
												}}
												onMouseOver={e => { if (!isFull) e.currentTarget.style.opacity = '0.9' }}
												onMouseOut={e => { if (!isFull) e.currentTarget.style.opacity = '1' }}
											>
												{isFull ? 'Event Full' : 'Register Now'}
											</button>
										)}
									</div>
								</div>
							)
						})}
					</div>
				)}
			</div>
		</div>
	)
}
