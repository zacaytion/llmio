-- +goose Up
-- +goose StatementBegin

-- Memberships table: relationship between users and groups
-- Features:
--   - Role-based access (admin/member)
--   - Pending invitations (accepted_at IS NULL)
--   - Inviter tracking for audit
--   - Last-admin protection via database trigger

CREATE TABLE memberships (
    id              BIGSERIAL PRIMARY KEY,
    group_id        BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            TEXT NOT NULL DEFAULT 'member',
    inviter_id      BIGINT NOT NULL REFERENCES users(id),
    accepted_at     TIMESTAMPTZ,  -- NULL = pending invitation

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT memberships_role_valid
        CHECK (role IN ('admin', 'member')),
    CONSTRAINT memberships_unique_user_group
        UNIQUE (group_id, user_id)
);

-- Indexes for common queries
CREATE INDEX memberships_user_id_idx ON memberships(user_id);
CREATE INDEX memberships_group_id_idx ON memberships(group_id);
CREATE INDEX memberships_inviter_id_idx ON memberships(inviter_id);
-- Partial index for pending invitations
CREATE INDEX memberships_pending_idx ON memberships(user_id, accepted_at)
    WHERE accepted_at IS NULL;
-- Index for admin counting queries
CREATE INDEX memberships_group_role_idx ON memberships(group_id, role)
    WHERE role = 'admin' AND accepted_at IS NOT NULL;

-- Updated at trigger
CREATE TRIGGER memberships_updated_at
    BEFORE UPDATE ON memberships
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Last-admin protection trigger
-- Prevents removing or demoting the last administrator of a group
-- This is enforced at the database level per TC-005 to prevent race conditions
CREATE OR REPLACE FUNCTION prevent_last_admin_removal()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    admin_count INTEGER;
BEGIN
    -- Only check for admin role changes or deletions of admins
    IF (TG_OP = 'DELETE' AND OLD.role = 'admin' AND OLD.accepted_at IS NOT NULL) OR
       (TG_OP = 'UPDATE' AND OLD.role = 'admin' AND NEW.role != 'admin' AND OLD.accepted_at IS NOT NULL) THEN

        -- Count remaining active admins (excluding the record being modified)
        SELECT COUNT(*) INTO admin_count
        FROM memberships
        WHERE group_id = OLD.group_id
          AND role = 'admin'
          AND accepted_at IS NOT NULL
          AND id != OLD.id;

        IF admin_count = 0 THEN
            RAISE EXCEPTION 'Cannot remove or demote the last administrator of a group'
                USING ERRCODE = 'P0001';  -- raise_exception
        END IF;
    END IF;

    RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;
END;
$$;

CREATE TRIGGER memberships_last_admin_protection
    BEFORE UPDATE OR DELETE ON memberships
    FOR EACH ROW
    EXECUTE FUNCTION prevent_last_admin_removal();

COMMENT ON TABLE memberships IS 'User-group relationships with role and invitation status';
COMMENT ON COLUMN memberships.role IS 'Either admin or member';
COMMENT ON COLUMN memberships.accepted_at IS 'NULL means pending invitation; non-null means active member';
COMMENT ON COLUMN memberships.inviter_id IS 'User who created this membership/invitation';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS memberships_last_admin_protection ON memberships;
DROP FUNCTION IF EXISTS prevent_last_admin_removal();
DROP TRIGGER IF EXISTS memberships_updated_at ON memberships;
DROP TABLE IF EXISTS memberships;

-- +goose StatementEnd
