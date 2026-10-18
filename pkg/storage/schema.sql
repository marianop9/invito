CREATE TABLE IF NOT EXISTS invitations (
    slug TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    data_json TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS rsvps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invitation_slug TEXT NOT NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    attending BOOLEAN NOT NULL,
    guest_count INTEGER NOT NULL DEFAULT 1,
    dietary_needs TEXT,
    song_request TEXT,
    personal_message TEXT,
    submitted_at DATETIME NOT NULL,
    FOREIGN KEY (invitation_slug) REFERENCES invitations(slug) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_rsvps_invitation_slug ON rsvps(invitation_slug);
