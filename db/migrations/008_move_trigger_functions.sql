-- +goose Up
-- +goose StatementBegin

-- Move trigger functions from public to private schema
-- These are internal implementation details that should not be directly called

-- 1. Create the functions in private schema (with same logic)

-- Updated at trigger function
CREATE OR REPLACE FUNCTION private.set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION private.set_updated_at() IS 'Trigger function to auto-update updated_at timestamp';

-- Last-admin protection trigger function
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

COMMENT ON FUNCTION private.protect_last_admin() IS 'Trigger function to prevent removing the last admin from a group';

-- 2. Update triggers to use private schema functions

-- Update users trigger
DROP TRIGGER IF EXISTS users_updated_at ON users;
CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION private.set_updated_at();

-- Update groups trigger
DROP TRIGGER IF EXISTS groups_updated_at ON groups;
CREATE TRIGGER groups_updated_at
    BEFORE UPDATE ON groups
    FOR EACH ROW
    EXECUTE FUNCTION private.set_updated_at();

-- Update memberships triggers
DROP TRIGGER IF EXISTS memberships_updated_at ON memberships;
CREATE TRIGGER memberships_updated_at
    BEFORE UPDATE ON memberships
    FOR EACH ROW
    EXECUTE FUNCTION private.set_updated_at();

DROP TRIGGER IF EXISTS memberships_last_admin_protection ON memberships;
CREATE TRIGGER memberships_last_admin_protection
    BEFORE UPDATE OR DELETE ON memberships
    FOR EACH ROW
    EXECUTE FUNCTION private.protect_last_admin();

-- 3. Drop old functions from public schema
DROP FUNCTION IF EXISTS public.update_updated_at() CASCADE;
DROP FUNCTION IF EXISTS public.prevent_last_admin_removal() CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore functions to public schema

-- Recreate update_updated_at in public
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate prevent_last_admin_removal in public
CREATE OR REPLACE FUNCTION prevent_last_admin_removal()
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

-- Restore triggers to use public schema functions
DROP TRIGGER IF EXISTS users_updated_at ON users;
CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS groups_updated_at ON groups;
CREATE TRIGGER groups_updated_at
    BEFORE UPDATE ON groups
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS memberships_updated_at ON memberships;
CREATE TRIGGER memberships_updated_at
    BEFORE UPDATE ON memberships
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS memberships_last_admin_protection ON memberships;
CREATE TRIGGER memberships_last_admin_protection
    BEFORE UPDATE OR DELETE ON memberships
    FOR EACH ROW
    EXECUTE FUNCTION prevent_last_admin_removal();

-- Drop private schema functions
DROP FUNCTION IF EXISTS private.set_updated_at() CASCADE;
DROP FUNCTION IF EXISTS private.protect_last_admin() CASCADE;

-- +goose StatementEnd
