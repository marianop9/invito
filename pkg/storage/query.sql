-- name: CountInvitations :one
SELECT COUNT(*) FROM invitations;

-- name: GetInvitationJSON :one
SELECT data_json FROM invitations
WHERE slug = ?;

-- name: ListInvitationsJSON :many
SELECT data_json FROM invitations
ORDER BY created_at ASC;

-- name: SaveInvitation :exec
INSERT INTO invitations (slug, title, description, data_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(slug) DO UPDATE SET
    title = excluded.title,
    description = excluded.description,
    data_json = excluded.data_json,
    updated_at = excluded.updated_at;

-- name: DeleteInvitation :execrows
DELETE FROM invitations
WHERE slug = ?;

-- name: InvitationExists :one
SELECT 1 FROM invitations
WHERE slug = ?;

-- name: CreateRSVP :exec
INSERT INTO rsvps (
    invitation_slug, name, email, attending, guest_count, dietary_needs, song_request, personal_message, submitted_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListRSVPs :many
SELECT name, email, attending, guest_count, dietary_needs, song_request, personal_message, submitted_at
FROM rsvps
WHERE invitation_slug = ?
ORDER BY submitted_at DESC;

-- name: GetRSVPStats :one
SELECT 
    COUNT(*) AS total_responses,
    CAST(COALESCE(SUM(CASE WHEN attending THEN 1 ELSE 0 END), 0) AS INTEGER) AS attending_count,
    CAST(COALESCE(SUM(CASE WHEN NOT attending THEN 1 ELSE 0 END), 0) AS INTEGER) AS declined_count,
    CAST(COALESCE(SUM(CASE WHEN attending THEN guest_count ELSE 0 END), 0) AS INTEGER) AS total_guests,
    CAST(COALESCE(SUM(CASE WHEN dietary_needs IS NOT NULL AND TRIM(dietary_needs) != '' THEN 1 ELSE 0 END), 0) AS INTEGER) AS dietary_count
FROM rsvps
WHERE invitation_slug = ?;
