-- Wrestling schema
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS elements (
    id                  integer PRIMARY KEY,
    name                text NOT NULL,
    icon                text NOT NULL DEFAULT '',
    category            text NOT NULL,
    subcategory         text,
    difficulty          text NOT NULL,
    difficulty_label    text NOT NULL DEFAULT '',
    reps                text NOT NULL DEFAULT '',
    muscles             text[] NOT NULL DEFAULT '{}',
    image               text NOT NULL,
    video               text,
    video_search        text,
    description         text NOT NULL DEFAULT '',
    requirements        jsonb NOT NULL DEFAULT '[]'::jsonb,
    steps               text[] NOT NULL DEFAULT '{}',
    tips                text NOT NULL DEFAULT '',
    demo_available     boolean NOT NULL DEFAULT false,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bars (
    id              integer PRIMARY KEY,
    name            text NOT NULL,
    icon            text NOT NULL DEFAULT '',
    type            text NOT NULL,
    description     text NOT NULL DEFAULT '',
    features        text[] NOT NULL DEFAULT '{}',
    tips            text NOT NULL DEFAULT '',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS base_blocks (
    id          bigserial PRIMARY KEY,
    title       text NOT NULL,
    content     text,
    items       text[] NOT NULL DEFAULT '{}',
    sort_order  integer NOT NULL DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS safety_items (
    id              integer PRIMARY KEY,
    name            text NOT NULL,
    icon            text NOT NULL DEFAULT '',
    category        text NOT NULL,
    image           text NOT NULL DEFAULT '',
    description     text NOT NULL DEFAULT '',
    benefits        text[] NOT NULL DEFAULT '{}',
    tips            text NOT NULL DEFAULT '',
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_elements_category ON elements(category);
CREATE INDEX IF NOT EXISTS idx_elements_subcategory ON elements(subcategory);
CREATE INDEX IF NOT EXISTS idx_elements_difficulty ON elements(difficulty);
CREATE INDEX IF NOT EXISTS idx_elements_name_trgm ON elements USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_elements_description_trgm ON elements USING gin (description gin_trgm_ops);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_elements_updated_at ON elements;
CREATE TRIGGER trg_elements_updated_at BEFORE UPDATE ON elements
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_bars_updated_at ON bars;
CREATE TRIGGER trg_bars_updated_at BEFORE UPDATE ON bars
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_base_blocks_updated_at ON base_blocks;
CREATE TRIGGER trg_base_blocks_updated_at BEFORE UPDATE ON base_blocks
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_safety_items_updated_at ON safety_items;
CREATE TRIGGER trg_safety_items_updated_at BEFORE UPDATE ON safety_items
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS schema_migrations (
    version bigint PRIMARY KEY,
    name text NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT now()
);
