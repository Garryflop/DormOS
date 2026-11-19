# Frontend — Ernar

React frontend for the **DormOS** dormitory management system.

## 🖥 Stack

| Technology | Usage |
|---|---|
| React 18 + TypeScript | UI framework |
| Vite | Build tool |
| React Router v6 | Client-side routing |
| Axios | HTTP client |
| Tailwind CSS | Utility styles |
| Inter Variable | Typography (Linear design system) |

## 📁 Structure

```
frontend/src/
├── api.ts                  # Axios client + all API calls
├── AuthContext.tsx         # Auth state (user, login, logout)
├── App.tsx                 # Router + protected routes
├── components/
│   └── Layout.tsx          # Sidebar + main layout
└── pages/
    ├── LoginPage.tsx       # Sign in
    ├── DashboardPage.tsx   # Overview + stats
    ├── IssuesPage.tsx      # Issues list + create + comments
    ├── DocumentsPage.tsx   # Documents
    ├── ActivitiesPage.tsx  # Activities + points
    └── AdminPage.tsx       # Admin panel (manager/admin only)
```

## 📄 Pages

| Page | Route | Access | Description |
|---|---|---|---|
| Login | `/login` | Public | Sign in with email + password |
| Dashboard | `/dashboard` | All | Stats overview, recent issues |
| Issues | `/issues` | All | Create/view issues, comments, status filter |
| Documents | `/documents` | All | Dormitory documents |
| Activities | `/activities` | All | Events and points |
| Admin | `/admin` | Manager+ | Manage all issues, update statuses |

## 🎨 Design

Based on the **Linear design system**:
- Dark-first: `#08090a` background
- Inter Variable with `cv01`, `ss03` OpenType features
- Brand indigo accent: `#5e6ad2` / `#7170ff`
- Semi-transparent white borders: `rgba(255,255,255,0.08)`
- Weight 510 as the signature UI weight

## 🚀 Running

```bash
cd frontend
npm install
npm run dev
```

App runs at `http://localhost:5173`

## 🔧 Environment

Create `.env` in `frontend/`:

```env
VITE_API_URL=http://localhost:8080/api/v1
```

## 🔌 API Endpoints Used

| Method | Endpoint | Description |
|---|---|---|
| POST | `/auth/login` | Login |
| GET | `/auth/me` | Get current user |
| GET | `/issues/my` | My issues |
| GET | `/issues` | All issues (manager+) |
| POST | `/issues` | Create issue |
| GET | `/issues/:id/comments` | Get comments |
| POST | `/issues/:id/comments` | Add comment |
| PATCH | `/issues/:id/status` | Update status |
| GET | `/categories` | List categories |

## 👤 Author

**Ernar** — Issue & Maintenance Service + React Frontend