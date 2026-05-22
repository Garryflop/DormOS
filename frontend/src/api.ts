import axios from 'axios'

const api = axios.create({
    baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
    headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
    const token = localStorage.getItem('token')
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
})

api.interceptors.response.use(
    (res) => res,
    (err) => {
        if (err.response?.status === 401) {
            localStorage.removeItem('token')
            window.location.href = '/login'
        }
        return Promise.reject(err)
    }
)

export default api

export const authAPI = {
    login: (email: string, password: string) => api.post('/auth/login', { email, password }),
    register: (data: RegisterDto) => api.post('/auth/register', data),
    me: () => api.get('/auth/profile'),
}

export const issuesAPI = {
    create: (data: CreateIssueDto) => api.post('/issues', data),
    getById: (id: string) => api.get(`/issues/${id}`),
    listMy: () => api.get('/issues/my'),
    listAll: () => api.get('/issues'),
    updateStatus: (id: string, status: string) => api.patch(`/issues/${id}/status`, { status }),
    delete: (id: string) => api.delete(`/issues/${id}`),
    addComment: (id: string, text: string) => api.post(`/issues/${id}/comments`, { text }),
    listComments: (id: string) => api.get(`/issues/${id}/comments`),
    assignWorker: (id: string, worker_id: string) => api.patch(`/issues/${id}/assign`, { worker_id }),
    listWorkers: () => api.get('/issues/workers'),
    listCategories: () => api.get('/categories'),
}

export const roomsAPI = {
    list: (floor?: number) => api.get('/rooms', { params: floor ? { floor } : {} }),
    get: (id: string) => api.get(`/rooms/${id}`),
    create: (data: { room_number: string; floor: number; capacity: number }) => api.post('/rooms', data),
}

export const residentsAPI = {
    list: (role?: string) => api.get('/residents', { params: role ? { role } : {} }),
    get: (userId: string) => api.get(`/residents/${userId}`),
    assign: (userId: string, roomId: string) => api.post('/residents', { user_id: userId, room_id: roomId }),
    remove: (userId: string) => api.delete(`/residents/${userId}`),
    updateRole: (userId: string, role: string) => api.patch(`/residents/${userId}/role`, { role }),
}

export const filesAPI = {
    upload: (file: File) => {
        const formData = new FormData()
        formData.append('file', file)
        return api.post('/files/upload', formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
        })
    },
    getUrl: (id: string) => api.get(`/files/${id}/url`),
}

export const eventsAPI = {
    list: () => api.get('/events'),
    create: (data: { title: string; description: string; location: string; start_time: number; end_time: number; max_participants: number }) => api.post('/events', data),
    update: (id: string, data: { title: string; description: string; location: string; start_time: number; end_time: number; max_participants: number }) => api.put(`/events/${id}`, data),
    delete: (id: string) => api.delete(`/events/${id}`),
    register: (id: string) => api.post(`/events/${id}/register`),
    cancel: (id: string) => api.post(`/events/${id}/cancel`),
}

export const notificationsAPI = {
    list: (userId: string) => api.get('/notifications', { params: { user_id: userId } }),
    markAsRead: (id: string) => api.patch(`/notifications/${id}/read`),
    markAllAsRead: (userId: string) => api.post('/notifications/read-all', { user_id: userId }),
    send: (data: { user_id: string; title: string; message: string; channel: string }) => api.post('/notifications', data),
}

export interface RegisterDto {
    full_name: string
    email: string
    password: string
}

export interface CreateIssueDto {
    room_number: string
    category_id: string
    title: string
    description: string
    photo_url?: string
}

export interface Issue {
    id: string
    user_id: string
    room_number: string
    category_id: string
    title: string
    description: string
    status: 'open' | 'in_progress' | 'resolved' | 'closed'
    worker_id?: string
    photo_url?: string
    created_at: string
    updated_at: string
}

export interface Category { id: string; name: string }
export interface Worker { id: string; name: string; specialty: string }
export interface Comment { id: string; issue_id: string; user_id: string; content: string; created_at: string }