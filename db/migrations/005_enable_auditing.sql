-- +goose Up
-- +goose StatementBegin

-- Enable audit triggers on groups and memberships tables
-- This is a separate migration to ensure tables exist before triggers are attached

-- Audit trigger for groups table
CREATE TRIGGER groups_audit
    AFTER INSERT OR UPDATE OR DELETE ON groups
    FOR EACH ROW
    EXECUTE FUNCTION audit.insert_update_delete_trigger();

-- Audit trigger for memberships table
CREATE TRIGGER memberships_audit
    AFTER INSERT OR UPDATE OR DELETE ON memberships
    FOR EACH ROW
    EXECUTE FUNCTION audit.insert_update_delete_trigger();

COMMENT ON TRIGGER groups_audit ON groups IS 'Captures all changes to groups in audit.record_version';
COMMENT ON TRIGGER memberships_audit ON memberships IS 'Captures all changes to memberships in audit.record_version';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS memberships_audit ON memberships;
DROP TRIGGER IF EXISTS groups_audit ON groups;

-- +goose StatementEnd
