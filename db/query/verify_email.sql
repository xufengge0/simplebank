-- name: CreateVerifyEmail :one
INSERT INTO verify_emails(
    username,
    email,
    secret_code
) VALUES(
    $1,$2,$3
) RETURNING *;

-- name: UpdateVerifyEmail :one
UPDATE verify_emails
    SET is_used = true
    WHERE id = @id
    and secret_code = @secret_code
    and is_used = false
    and expired_at > now()
RETURNING *;