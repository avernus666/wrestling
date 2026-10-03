-- Wrestling hardening: account security, richer training records and auditability.
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified boolean NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_count integer NOT NULL DEFAULT 0 CHECK (failed_login_count >= 0);
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until timestamptz;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at timestamptz;

ALTER TABLE sessions ADD COLUMN IF NOT EXISTS revoked_at timestamptz;
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS user_agent text NOT NULL DEFAULT '';
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS ip_address inet;
CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(user_id, expires_at) WHERE revoked_at IS NULL;

ALTER TABLE workout_sessions ADD COLUMN IF NOT EXISTS notes text NOT NULL DEFAULT '';
ALTER TABLE workout_sessions ADD COLUMN IF NOT EXISTS perceived_effort smallint CHECK (perceived_effort IS NULL OR perceived_effort BETWEEN 1 AND 10);
ALTER TABLE workout_sessions ADD COLUMN IF NOT EXISTS calories integer CHECK (calories IS NULL OR calories >= 0);

CREATE TABLE IF NOT EXISTS training_plans (
 id uuid PRIMARY KEY,
 name varchar(160) NOT NULL,
 description varchar(3000) NOT NULL DEFAULT '',
 stage smallint NOT NULL CHECK(stage BETWEEN 1 AND 3),
 duration_minutes integer NOT NULL CHECK(duration_minutes BETWEEN 5 AND 180),
 equipment text[] NOT NULL DEFAULT '{}',
 active boolean NOT NULL DEFAULT true,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS training_plan_items (
 plan_id uuid NOT NULL REFERENCES training_plans(id) ON DELETE CASCADE,
 element_id integer NOT NULL REFERENCES elements(id) ON DELETE RESTRICT,
 position integer NOT NULL CHECK(position > 0),
 sets integer CHECK(sets IS NULL OR sets BETWEEN 1 AND 50),
 reps text NOT NULL DEFAULT '',
 rest_seconds integer CHECK(rest_seconds IS NULL OR rest_seconds BETWEEN 0 AND 900),
 PRIMARY KEY(plan_id, position)
);
CREATE INDEX IF NOT EXISTS idx_training_plans_stage ON training_plans(stage, active);

CREATE TABLE IF NOT EXISTS readiness_checks (
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 check_date date NOT NULL DEFAULT CURRENT_DATE,
 pain_free boolean NOT NULL,
 surface_safe boolean NOT NULL,
 equipment_safe boolean NOT NULL,
 space_clear boolean NOT NULL,
 hydrated boolean NOT NULL,
 recovery_ready boolean NOT NULL,
 notes text NOT NULL DEFAULT '',
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(user_id, check_date)
);

CREATE TABLE IF NOT EXISTS audit_events (
 id bigserial PRIMARY KEY,
 user_id uuid REFERENCES users(id) ON DELETE SET NULL,
 request_id text NOT NULL DEFAULT '',
 action varchar(80) NOT NULL,
 entity_type varchar(80) NOT NULL DEFAULT '',
 entity_id varchar(120) NOT NULL DEFAULT '',
 ip_address inet,
 user_agent text NOT NULL DEFAULT '',
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_user_created ON audit_events(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_action_created ON audit_events(action, created_at DESC);

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
