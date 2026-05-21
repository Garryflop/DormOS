import { FileText } from 'lucide-react'

export default function DocumentsPage() {
    return (
        <div style={{ maxWidth: '800px' }}>
            <div style={{ marginBottom: '28px' }}>
                <h1 style={{ fontSize: '20px', fontWeight: 510, letterSpacing: '-0.24px', marginBottom: '4px' }}>Documents</h1>
                <p style={{ fontSize: '13px', color: 'var(--text4)' }}>Your dormitory documents and certificates</p>
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
