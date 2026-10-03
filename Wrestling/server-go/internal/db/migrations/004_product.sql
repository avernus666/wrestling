CREATE TABLE IF NOT EXISTS users (
 id uuid PRIMARY KEY,
 email text NOT NULL UNIQUE,
 password_hash text NOT NULL,
 display_name text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS sessions (
 id uuid PRIMARY KEY,
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 token_hash text NOT NULL UNIQUE,
 expires_at timestamptz NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 last_seen_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS user_progress (
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 element_id integer NOT NULL REFERENCES elements(id) ON DELETE CASCADE,
 completed boolean NOT NULL DEFAULT false,
 favorite boolean NOT NULL DEFAULT false,
 updated_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(user_id, element_id)
);

CREATE TABLE IF NOT EXISTS workout_sessions (
 id uuid PRIMARY KEY,
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 goal text NOT NULL,
 duration_minutes integer NOT NULL CHECK(duration_minutes BETWEEN 5 AND 180),
 equipment text NOT NULL,
 started_at timestamptz NOT NULL,
 completed_at timestamptz,
 total_exercises integer NOT NULL DEFAULT 0 CHECK(total_exercises >= 0),
 completed_exercises integer NOT NULL DEFAULT 0 CHECK(completed_exercises >= 0),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_workout_sessions_user_date ON workout_sessions(user_id, started_at DESC);

CREATE TABLE IF NOT EXISTS workout_session_exercises (
 session_id uuid NOT NULL REFERENCES workout_sessions(id) ON DELETE CASCADE,
 element_id integer NOT NULL REFERENCES elements(id),
 position integer NOT NULL CHECK(position > 0),
 completed boolean NOT NULL DEFAULT false,
 PRIMARY KEY(session_id, element_id)
);

CREATE TABLE IF NOT EXISTS user_skills (
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 skill_key text NOT NULL,
 level integer NOT NULL DEFAULT 0 CHECK(level >= 0),
 updated_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(user_id, skill_key)
);

CREATE OR REPLACE FUNCTION cleanup_expired_sessions() RETURNS void LANGUAGE SQL AS $$
 DELETE FROM sessions WHERE expires_at < now();
$$;
