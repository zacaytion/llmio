-- +goose Up
-- +goose StatementBegin

-- Move application tables from public to data schema
-- This provides clean separation between application data and system objects

-- Move tables (foreign keys and indexes move automatically with the table)
ALTER TABLE public.users SET SCHEMA data;
ALTER TABLE public.groups SET SCHEMA data;
ALTER TABLE public.memberships SET SCHEMA data;

-- Update trigger function references to use fully qualified table names
-- The protect_last_admin function queries memberships table directly

-- Recreate protect_last_admin with data schema reference
CREATE OR REPLACE FUNCTION private.protect_last_admin()
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
        -- Note: Uses data.memberships since tables have moved
        SELECT COUNT(*) INTO admin_count
        FROM data.memberships
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

-- Update audit trigger function to handle new schema
-- The audit function already uses TG_TABLE_SCHEMA which will reflect 'data'

-- Recreate triggers on moved tables (triggers don't automatically move)
-- Note: Triggers need to be recreated with fully qualified function names

-- Users table triggers
DROP TRIGGER IF EXISTS users_updated_at ON data.users;
CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON data.users
    FOR EACH ROW
    EXECUTE FUNCTION private.set_updated_at();

-- Groups table triggers
DROP TRIGGER IF EXISTS groups_updated_at ON data.groups;
CREATE TRIGGER groups_updated_at
    BEFORE UPDATE ON data.groups
    FOR EACH ROW
    EXECUTE FUNCTION private.set_updated_at();

DROP TRIGGER IF EXISTS groups_audit ON data.groups;
CREATE TRIGGER groups_audit
    AFTER INSERT OR UPDATE OR DELETE ON data.groups
    FOR EACH ROW
    EXECUTE FUNCTION audit.insert_update_delete_trigger();

-- Memberships table triggers
DROP TRIGGER IF EXISTS memberships_updated_at ON data.memberships;
CREATE TRIGGER memberships_updated_at
    BEFORE UPDATE ON data.memberships
    FOR EACH ROW
    EXECUTE FUNCTION private.set_updated_at();

DROP TRIGGER IF EXISTS memberships_last_admin_protection ON data.memberships;
CREATE TRIGGER memberships_last_admin_protection
    BEFORE UPDATE OR DELETE ON data.memberships
    FOR EACH ROW
    EXECUTE FUNCTION private.protect_last_admin();

DROP TRIGGER IF EXISTS memberships_audit ON data.memberships;
CREATE TRIGGER memberships_audit
    AFTER INSERT OR UPDATE OR DELETE ON data.memberships
    FOR EACH ROW
    EXECUTE FUNCTION audit.insert_update_delete_trigger();

-- Set search_path to include data schema for unqualified references
-- This allows existing code to work without modification during transition
--
-- We set it at the ROLE level for the postgres user, which persists across
-- all connections by that role. This is more permanent than session-level
-- settings and works regardless of which database is being connected to.
ALTER ROLE postgres SET search_path TO data, public;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Reset search_path to default for postgres role
ALTER ROLE postgres SET search_path TO public;

-- Move tables back to public schema
ALTER TABLE data.memberships SET SCHEMA public;
ALTER TABLE data.groups SET SCHEMA public;
ALTER TABLE data.users SET SCHEMA public;

-- Recreate protect_last_admin with public schema reference
CREATE OR REPLACE FUNCTION private.protect_last_admin()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    admin_count INTEGER;
BEGIN
    IF (TG_OP = 'DELETE' AND OLD.role = 'admin' AND OLD.accepted_at IS NOT NULL) OR
       (TG_OP = 'UPDATE' AND OLD.role = 'admin' AND NEW.role != 'admin' AND OLD.accepted_at IS NOT NULL) THEN

        SELECT COUNT(*) INTO admin_count
        FROM memberships
        WHERE group_id = OLD.group_id
          AND role = 'admin'
          AND accepted_at IS NOT NULL
          AND id != OLD.id;

        IF admin_count = 0 THEN
            RAISE EXCEPTION 'Cannot remove or demote the last administrator of a group'
                USING ERRCODE = 'P0001';
        END IF;
    END IF;

    RETURN CASE WHEN TG_OP = 'DELETE' THEN OLD ELSE NEW END;
END;
$$;

-- Recreate triggers on public schema tables
DROP TRIGGER IF EXISTS users_updated_at ON public.users;
CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON public.users
    FOR EACH ROW
    EXECUTE FUNCTION private.set_updated_at();

DROP TRIGGER IF EXISTS groups_updated_at ON public.groups;
CREATE TRIGGER groups_updated_at
    BEFORE UPDATE ON public.groups
    FOR EACH ROW
    EXECUTE FUNCTION private.set_updated_at();

DROP TRIGGER IF EXISTS groups_audit ON public.groups;
CREATE TRIGGER groups_audit
    AFTER INSERT OR UPDATE OR DELETE ON public.groups
    FOR EACH ROW
    EXECUTE FUNCTION audit.insert_update_delete_trigger();

DROP TRIGGER IF EXISTS memberships_updated_at ON public.memberships;
CREATE TRIGGER memberships_updated_at
    BEFORE UPDATE ON public.memberships
    FOR EACH ROW
    EXECUTE FUNCTION private.set_updated_at();

DROP TRIGGER IF EXISTS memberships_last_admin_protection ON public.memberships;
CREATE TRIGGER memberships_last_admin_protection
    BEFORE UPDATE OR DELETE ON public.memberships
    FOR EACH ROW
    EXECUTE FUNCTION private.protect_last_admin();

DROP TRIGGER IF EXISTS memberships_audit ON public.memberships;
CREATE TRIGGER memberships_audit
    AFTER INSERT OR UPDATE OR DELETE ON public.memberships
    FOR EACH ROW
    EXECUTE FUNCTION audit.insert_update_delete_trigger();

-- +goose StatementEnd
