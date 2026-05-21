CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS issue_categories (
                                                id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                                                name       VARCHAR(100) NOT NULL UNIQUE,
                                                created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO issue_categories (name) VALUES
                                        ('Plumber'),
                                        ('Electrician'),
                                        ('Furniture'),
                                        ('Heating'),
                                        ('Network'),
                                        ('Security'),
                                        ('Other')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS workers (
                                       id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                                       name       VARCHAR(255) NOT NULL,
                                       specialty  VARCHAR(100) NOT NULL,
                                       phone      VARCHAR(30),
                                       is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
                                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS issues (
                                      id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
                                      user_id     UUID         NOT NULL,
                                      room_number VARCHAR(20)  NOT NULL,
                                      category_id UUID         NOT NULL REFERENCES issue_categories(id),
                                      title       VARCHAR(255) NOT NULL,
                                      description TEXT         NOT NULL,
                                      status      VARCHAR(50)  NOT NULL DEFAULT 'open'
                                          CHECK (status IN ('open', 'in_progress', 'resolved', 'closed')),
                                      worker_id   UUID         NULL REFERENCES workers(id),
                                      photo_url   VARCHAR(1024) NULL,
                                      created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
                                      updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_issues_user_id     ON issues(user_id);
CREATE INDEX idx_issues_status      ON issues(status);
CREATE INDEX idx_issues_room_number ON issues(room_number);

CREATE TABLE IF NOT EXISTS issue_comments (
                                              id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
                                              issue_id   UUID        NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
                                              user_id    UUID        NOT NULL,
                                              text       TEXT        NOT NULL,
                                              created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_issue_comments_issue_id ON issue_comments(issue_id);
