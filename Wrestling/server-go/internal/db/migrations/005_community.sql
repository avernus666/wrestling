CREATE TABLE IF NOT EXISTS community_elements (
  id UUID PRIMARY KEY,
  author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name VARCHAR(120) NOT NULL,
  description VARCHAR(3000) NOT NULL,
  category VARCHAR(20) NOT NULL CHECK (category IN ('стойка','партер','офп')),
  direction VARCHAR(80) NOT NULL,
  youtube_url VARCHAR(2048) NOT NULL,
  youtube_video_id VARCHAR(32) NOT NULL,
  moderation_status VARCHAR(20) NOT NULL DEFAULT 'approved' CHECK (moderation_status IN ('approved','pending','rejected')),
  moderation_reason VARCHAR(500) NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_community_elements_category_created ON community_elements(category, created_at DESC) WHERE moderation_status='approved';
CREATE INDEX IF NOT EXISTS idx_community_elements_author ON community_elements(author_id, created_at DESC);
CREATE TABLE IF NOT EXISTS community_comments (
  id UUID PRIMARY KEY,
  element_id UUID NOT NULL REFERENCES community_elements(id) ON DELETE CASCADE,
  author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  body VARCHAR(1200) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_community_comments_element_created ON community_comments(element_id, created_at ASC);
CREATE TABLE IF NOT EXISTS community_reports (
  id UUID PRIMARY KEY,
  element_id UUID NOT NULL REFERENCES community_elements(id) ON DELETE CASCADE,
  reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  reason VARCHAR(80) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(element_id, reporter_id)
);
