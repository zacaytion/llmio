-- +goose Up
-- +goose StatementBegin

-- Groups table: organizational containers for collaborative decision-making
-- Features:
--   - 11 permission flags for fine-grained access control
--   - Optional parent_id for subgroup hierarchy
--   - Case-insensitive unique handles via CITEXT
--   - Soft deletion via archived_at timestamp

CREATE TABLE groups (
    id                                  BIGSERIAL PRIMARY KEY,
    name                                TEXT NOT NULL,
    handle                              CITEXT NOT NULL,
    description                         TEXT,
    parent_id                           BIGINT REFERENCES groups(id) ON DELETE SET NULL,
    created_by_id                       BIGINT NOT NULL REFERENCES users(id),
    archived_at                         TIMESTAMPTZ,

    -- Permission flags (all BOOLEAN NOT NULL with defaults)
    members_can_add_members             BOOLEAN NOT NULL DEFAULT TRUE,
    members_can_add_guests              BOOLEAN NOT NULL DEFAULT TRUE,
    members_can_start_discussions       BOOLEAN NOT NULL DEFAULT TRUE,
    members_can_raise_motions           BOOLEAN NOT NULL DEFAULT TRUE,
    members_can_edit_discussions        BOOLEAN NOT NULL DEFAULT FALSE,
    members_can_edit_comments           BOOLEAN NOT NULL DEFAULT TRUE,
    members_can_delete_comments         BOOLEAN NOT NULL DEFAULT TRUE,
    members_can_announce                BOOLEAN NOT NULL DEFAULT FALSE,
    members_can_create_subgroups        BOOLEAN NOT NULL DEFAULT FALSE,
    admins_can_edit_user_content        BOOLEAN NOT NULL DEFAULT FALSE,
    parent_members_can_see_discussions  BOOLEAN NOT NULL DEFAULT FALSE,

    -- Timestamps
    created_at                          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT groups_name_length
        CHECK (LENGTH(name) BETWEEN 1 AND 255),
    CONSTRAINT groups_handle_format
        CHECK (handle ~* '^[a-z0-9][a-z0-9-]*[a-z0-9]$' AND LENGTH(handle) BETWEEN 3 AND 100),
    CONSTRAINT groups_parent_not_self
        CHECK (parent_id IS NULL OR parent_id != id)
);

-- Unique constraint on handle (CITEXT handles case-insensitivity automatically)
CREATE UNIQUE INDEX groups_handle_key ON groups(handle);

-- Indexes for common queries
CREATE INDEX groups_parent_id_idx ON groups(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX groups_created_by_id_idx ON groups(created_by_id);
CREATE INDEX groups_archived_at_idx ON groups(archived_at) WHERE archived_at IS NOT NULL;
CREATE INDEX groups_created_at_idx ON groups(created_at);

-- Updated at trigger (reuse function from users migration)
CREATE TRIGGER groups_updated_at
    BEFORE UPDATE ON groups
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

COMMENT ON TABLE groups IS 'Organizational containers with permission-based membership';
COMMENT ON COLUMN groups.handle IS 'URL-safe identifier, case-insensitive unique';
COMMENT ON COLUMN groups.parent_id IS 'FK to parent group for hierarchy support';
COMMENT ON COLUMN groups.archived_at IS 'Soft deletion timestamp; non-null means archived';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS groups_updated_at ON groups;
DROP TABLE IF EXISTS groups;

-- +goose StatementEnd
