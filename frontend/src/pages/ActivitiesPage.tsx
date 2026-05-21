import { Star, Target, Trophy } from 'lucide-react'

export default function ActivitiesPage() {
    return (
        <div style={{ maxWidth: '800px' }}>
            <div style={{ marginBottom: '28px' }}>
                <h1 style={{ fontSize: '20px', fontWeight: 510, letterSpacing: '-0.24px', marginBottom: '4px' }}>Activities</h1>
                <p style={{ fontSize: '13px', color: 'var(--text4)' }}>Dormitory events, points, and achievements</p>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px', marginBottom: '20px' }}>
                {[
                    { label: 'Total Points', value: '—', icon: Star, color: '#f59e0b' },
                    { label: 'Events Attended', value: '—', icon: Target, color: '#7170ff' },
                ].map(card => {
                    const Icon = card.icon
                    return (
                        <div key={card.label} style={{
                            background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(255,255,255,0.08)',
                            borderRadius: '8px', padding: '20px 24px',
                            display: 'flex', alignItems: 'center', gap: '16px',
                        }}>
                            <div style={{ 
                                background: `${card.color}12`, 
                                padding: '10px', 
                                borderRadius: '10px', 
                                color: card.color,
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center'
                            }}>
                                <Icon size={20} />
                            </div>
                            <div>
                                <div style={{ fontSize: '24px', fontWeight: 510, letterSpacing: '-0.288px', color: 'var(--text)' }}>
                                    {card.value}
                                </div>
                                <div style={{ fontSize: '12px', color: 'var(--text4)' }}>{card.label}</div>
                            </div>
                        </div>
                    )
                })}
            </div>

            <div style={{
                background: 'rgba(255,255,255,0.02)', border: '1px solid rgba(255,255,255,0.08)',
                borderRadius: '8px', padding: '48px', textAlign: 'center',
            }}>
                <div style={{ 
                    display: 'inline-flex', 
                    alignItems: 'center', 
                    justifyContent: 'center',
                    background: 'rgba(255,255,255,0.02)', 
                    border: '1px solid rgba(255,255,255,0.06)',
                    padding: '16px', 
                    borderRadius: '50%',
                    color: 'var(--accent)',
                    marginBottom: '16px' 
                }}>
                    <Trophy size={32} />
                </div>
                <div style={{ fontSize: '14px', fontWeight: 510, color: 'var(--text2)', marginBottom: '6px', letterSpacing: '-0.182px' }}>
                    No activities yet
                </div>
                <div style={{ fontSize: '13px', color: 'var(--text4)' }}>
                    Participate in dormitory events to earn points
                </div>
            </div>
        </div>
    )
}
