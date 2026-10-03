-- Wrestling database hardening and search acceleration.

ALTER TABLE elements
    ADD CONSTRAINT elements_category_nonempty CHECK (length(btrim(category)) > 0),
    ADD CONSTRAINT elements_name_nonempty CHECK (length(btrim(name)) > 0),
    ADD CONSTRAINT elements_difficulty_valid CHECK (difficulty IN ('easy', 'medium', 'hard'));

CREATE INDEX IF NOT EXISTS idx_elements_muscles_gin ON elements USING gin (muscles);
CREATE INDEX IF NOT EXISTS idx_elements_category_difficulty ON elements(category, difficulty);
CREATE INDEX IF NOT EXISTS idx_elements_subcategory_difficulty ON elements(subcategory, difficulty);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;
