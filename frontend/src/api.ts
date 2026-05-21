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
export interface Comment { id: string; issue_id: string; user_id: string; text: string; created_at: string }