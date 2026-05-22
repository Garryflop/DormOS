import { useAuth } from '../AuthContext'
import { FileText, Sparkles, Building2, AlertCircle } from 'lucide-react'

export default function DocumentsPage() {
    const { user } = useAuth()
    
    // Check if resident is assigned to a room
    const hasRoom = !!user?.room_number && user?.room_number !== 'Unassigned'
    const isStaff = user?.role === 'manager' || user?.role === 'admin'

    return (
        <div style={{ maxWidth: '800px' }}>
            <div style={{ marginBottom: '28px' }}>
                <h1 style={{ fontSize: '20px', fontWeight: 510, letterSpacing: '-0.24px', marginBottom: '4px' }}>Documents</h1>
                <p style={{ fontSize: '13px', color: 'var(--text4)' }}>Your dormitory documents and certificates</p>
            </div>

            {/* Resident Welcome & Celebration Alert */}
            {!isStaff && hasRoom && (
                <div style={{
                    background: 'linear-gradient(135deg, rgba(16, 185, 129, 0.08) 0%, rgba(5, 150, 105, 0.03) 100%)',
                    border: '1px solid rgba(16, 185, 129, 0.2)',
                    borderRadius: '12px',
                    padding: '24px',
                    marginBottom: '28px',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '20px',
                    position: 'relative',
                    overflow: 'hidden'
                }}>
                    <div style={{
                        position: 'absolute', top: '-10px', right: '-10px', opacity: 0.08, color: '#10b981'
                    }}>
                        <Sparkles size={120} />
                    </div>
                    <div style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        background: 'rgba(16, 185, 129, 0.15)',
                        border: '1px solid rgba(16, 185, 129, 0.3)',
                        width: '54px',
                        height: '54px',
                        borderRadius: '12px',
                        color: '#10b981',
                        flexShrink: 0
                    }}>
                        <Building2 size={26} />
                    </div>
                    <div style={{ zIndex: 1 }}>
                        <h2 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--text)', marginBottom: '4px', display: 'flex', alignItems: 'center', gap: '8px' }}>
                            Congratulations! You are officially a Resident <Sparkles size={16} color="#fbbf24" style={{ fill: '#fbbf24' }} />
                        </h2>
                        <p style={{ fontSize: '13.5px', color: 'var(--text3)', lineHeight: '1.5' }}>
                            Welcome to the dormitory community! You have been successfully assigned to <strong>Room {user?.room_number}</strong>. You now have full access to file maintenance requests, register for community activities, and view all resident-only notices.
                        </p>
                    </div>
                </div>
            )}

            {/* Unassigned Warning Alert */}
            {!isStaff && !hasRoom && (
                <div style={{
                    background: 'linear-gradient(135deg, rgba(245, 158, 11, 0.08) 0%, rgba(217, 119, 6, 0.03) 100%)',
                    border: '1px solid rgba(245, 158, 11, 0.2)',
                    borderRadius: '12px',
                    padding: '24px',
                    marginBottom: '28px',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '20px',
                    position: 'relative',
                    overflow: 'hidden'
                }}>
                    <div style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        background: 'rgba(245, 158, 11, 0.15)',
                        border: '1px solid rgba(245, 158, 11, 0.3)',
                        width: '54px',
                        height: '54px',
                        borderRadius: '12px',
                        color: '#f59e0b',
                        flexShrink: 0
                    }}>
                        <AlertCircle size={26} />
                    </div>
                    <div>
                        <h2 style={{ fontSize: '16px', fontWeight: 600, color: 'var(--text)', marginBottom: '4px' }}>
                            Awaiting Room Assignment
                        </h2>
                        <p style={{ fontSize: '13.5px', color: 'var(--text3)', lineHeight: '1.5' }}>
                            You have successfully registered. To officially become a resident and get access to submit maintenance requests, please contact the dormitory administrator to assign you a room.
                        </p>
                    </div>
                </div>
            )}

            {/* Standard documents list card */}
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
                    <FileText size={32} />
                </div>
                <div style={{ fontSize: '14px', fontWeight: 510, color: 'var(--text2)', marginBottom: '6px', letterSpacing: '-0.182px' }}>
                    No documents yet
                </div>
                <div style={{ fontSize: '13px', color: 'var(--text4)' }}>
                    Documents from the administration will appear here
                </div>
            </div>
        </div>
    )
}
