--
-- PostgreSQL database dump
--

\restrict 2pqiIIKKIQf0yIgtJbuSh2yjI6fskVVQdxPo7hAy63YEIaqGYh56i6D2ZB6YI49

-- Dumped from database version 18.1 (Debian 18.1-1.pgdg13+2)
-- Dumped by pg_dump version 18.1 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pghero; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA pghero;


--
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;


--
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';


--
-- Name: hstore; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS hstore WITH SCHEMA public;


--
-- Name: EXTENSION hstore; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION hstore IS 'data type for storing sets of (key, value) pairs';


--
-- Name: pg_stat_statements; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_stat_statements WITH SCHEMA public;


--
-- Name: EXTENSION pg_stat_statements; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pg_stat_statements IS 'track planning and execution statistics of all SQL statements executed';


--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: action_mailbox_inbound_emails; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.action_mailbox_inbound_emails (
    id bigint NOT NULL,
    status integer DEFAULT 0 NOT NULL,
    message_id character varying NOT NULL,
    message_checksum character varying NOT NULL,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: action_mailbox_inbound_emails_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.action_mailbox_inbound_emails_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: action_mailbox_inbound_emails_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.action_mailbox_inbound_emails_id_seq OWNED BY public.action_mailbox_inbound_emails.id;


--
-- Name: active_admin_comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.active_admin_comments (
    id bigint NOT NULL,
    namespace character varying,
    body text,
    resource_type character varying,
    resource_id bigint,
    author_type character varying,
    author_id bigint,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: active_admin_comments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.active_admin_comments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: active_admin_comments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.active_admin_comments_id_seq OWNED BY public.active_admin_comments.id;


--
-- Name: active_storage_attachments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.active_storage_attachments (
    id bigint NOT NULL,
    name character varying NOT NULL,
    record_type character varying NOT NULL,
    record_id bigint NOT NULL,
    blob_id bigint NOT NULL,
    created_at timestamp without time zone NOT NULL,
    group_id integer
);


--
-- Name: active_storage_attachments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.active_storage_attachments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: active_storage_attachments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.active_storage_attachments_id_seq OWNED BY public.active_storage_attachments.id;


--
-- Name: active_storage_blobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.active_storage_blobs (
    id bigint NOT NULL,
    key character varying NOT NULL,
    filename character varying NOT NULL,
    content_type character varying,
    metadata text,
    byte_size bigint NOT NULL,
    checksum character varying,
    created_at timestamp without time zone NOT NULL,
    service_name character varying NOT NULL
);


--
-- Name: active_storage_blobs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.active_storage_blobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: active_storage_blobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.active_storage_blobs_id_seq OWNED BY public.active_storage_blobs.id;


--
-- Name: active_storage_variant_records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.active_storage_variant_records (
    id bigint NOT NULL,
    blob_id bigint NOT NULL,
    variation_digest character varying NOT NULL
);


--
-- Name: active_storage_variant_records_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.active_storage_variant_records_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: active_storage_variant_records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.active_storage_variant_records_id_seq OWNED BY public.active_storage_variant_records.id;


--
-- Name: ar_internal_metadata; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ar_internal_metadata (
    key character varying NOT NULL,
    value character varying,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: attachments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attachments (
    id integer NOT NULL,
    user_id integer,
    filename character varying(255),
    location text,
    comment_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    filesize integer,
    file_file_name character varying,
    file_content_type character varying,
    file_file_size integer,
    file_updated_at timestamp without time zone,
    attachable_id integer,
    attachable_type character varying,
    migrated_to_document boolean DEFAULT false NOT NULL
);


--
-- Name: attachments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.attachments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: attachments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.attachments_id_seq OWNED BY public.attachments.id;


--
-- Name: blazer_audits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blazer_audits (
    id bigint NOT NULL,
    user_id bigint,
    query_id bigint,
    statement text,
    data_source character varying,
    created_at timestamp without time zone
);


--
-- Name: blazer_audits_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.blazer_audits_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: blazer_audits_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.blazer_audits_id_seq OWNED BY public.blazer_audits.id;


--
-- Name: blazer_checks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blazer_checks (
    id bigint NOT NULL,
    creator_id bigint,
    query_id bigint,
    state character varying,
    schedule character varying,
    emails text,
    slack_channels text,
    check_type character varying,
    message text,
    last_run_at timestamp without time zone,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: blazer_checks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.blazer_checks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: blazer_checks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.blazer_checks_id_seq OWNED BY public.blazer_checks.id;


--
-- Name: blazer_dashboard_queries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blazer_dashboard_queries (
    id bigint NOT NULL,
    dashboard_id bigint,
    query_id bigint,
    "position" integer,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: blazer_dashboard_queries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.blazer_dashboard_queries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: blazer_dashboard_queries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.blazer_dashboard_queries_id_seq OWNED BY public.blazer_dashboard_queries.id;


--
-- Name: blazer_dashboards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blazer_dashboards (
    id bigint NOT NULL,
    creator_id bigint,
    name character varying,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: blazer_dashboards_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.blazer_dashboards_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: blazer_dashboards_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.blazer_dashboards_id_seq OWNED BY public.blazer_dashboards.id;


--
-- Name: blazer_queries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blazer_queries (
    id bigint NOT NULL,
    creator_id bigint,
    name character varying,
    description text,
    statement text,
    data_source character varying,
    status character varying,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: blazer_queries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.blazer_queries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: blazer_queries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.blazer_queries_id_seq OWNED BY public.blazer_queries.id;


--
-- Name: blocked_domains; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blocked_domains (
    id bigint NOT NULL,
    name character varying
);


--
-- Name: blocked_domains_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.blocked_domains_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: blocked_domains_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.blocked_domains_id_seq OWNED BY public.blocked_domains.id;


--
-- Name: chatbots; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chatbots (
    id bigint NOT NULL,
    kind character varying,
    server character varying,
    channel character varying,
    access_token character varying,
    author_id integer,
    group_id integer,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL,
    name character varying,
    notification_only boolean DEFAULT false NOT NULL,
    webhook_kind character varying,
    event_kinds character varying[]
);


--
-- Name: chatbots_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.chatbots_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: chatbots_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.chatbots_id_seq OWNED BY public.chatbots.id;


--
-- Name: cohorts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cohorts (
    id integer NOT NULL,
    start_on date,
    end_on date
);


--
-- Name: cohorts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cohorts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: cohorts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cohorts_id_seq OWNED BY public.cohorts.id;


--
-- Name: comments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.comments (
    id integer NOT NULL,
    discussion_id integer DEFAULT 0,
    body text DEFAULT ''::text,
    user_id integer DEFAULT 0,
    parent_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    comment_votes_count integer DEFAULT 0 NOT NULL,
    attachments_count integer DEFAULT 0 NOT NULL,
    edited_at timestamp without time zone,
    versions_count integer DEFAULT 0,
    body_format character varying(10) DEFAULT 'md'::character varying NOT NULL,
    attachments jsonb DEFAULT '[]'::jsonb NOT NULL,
    discarded_at timestamp without time zone,
    discarded_by integer,
    content_locale character varying,
    link_previews jsonb DEFAULT '[]'::jsonb NOT NULL,
    parent_type character varying NOT NULL
);


--
-- Name: comments_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.comments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: comments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.comments_id_seq OWNED BY public.comments.id;


--
-- Name: default_group_covers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.default_group_covers (
    id integer NOT NULL,
    cover_photo_file_name character varying,
    cover_photo_content_type character varying,
    cover_photo_file_size integer,
    cover_photo_updated_at timestamp without time zone,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: default_group_covers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.default_group_covers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: default_group_covers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.default_group_covers_id_seq OWNED BY public.default_group_covers.id;


--
-- Name: demos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.demos (
    id bigint NOT NULL,
    author_id integer NOT NULL,
    group_id integer NOT NULL,
    name character varying NOT NULL,
    description character varying,
    recorded_at timestamp without time zone NOT NULL,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL,
    priority integer DEFAULT 0 NOT NULL,
    demo_handle character varying
);


--
-- Name: demos_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.demos_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: demos_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.demos_id_seq OWNED BY public.demos.id;


--
-- Name: discussion_readers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.discussion_readers (
    id integer NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    discussion_id integer NOT NULL,
    last_read_at timestamp without time zone,
    last_read_sequence_id integer DEFAULT 0 NOT NULL,
    volume integer DEFAULT 2 NOT NULL,
    participating boolean DEFAULT false NOT NULL,
    dismissed_at timestamp without time zone,
    read_ranges_string character varying,
    inviter_id integer,
    token character varying,
    revoked_at timestamp without time zone,
    admin boolean DEFAULT false NOT NULL,
    accepted_at timestamp without time zone,
    revoker_id integer,
    guest boolean DEFAULT false NOT NULL
);


--
-- Name: discussion_readers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.discussion_readers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: discussion_readers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.discussion_readers_id_seq OWNED BY public.discussion_readers.id;


--
-- Name: discussion_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.discussion_templates (
    id bigint NOT NULL,
    source_discussion_id integer,
    key character varying,
    group_id integer,
    "position" integer,
    author_id integer,
    title character varying,
    description text,
    description_format character varying(10) DEFAULT 'html'::character varying NOT NULL,
    process_name character varying,
    process_subtitle character varying,
    process_introduction character varying,
    process_introduction_format character varying DEFAULT 'html'::character varying NOT NULL,
    attachments jsonb DEFAULT '[]'::jsonb NOT NULL,
    max_depth integer DEFAULT 2 NOT NULL,
    newest_first boolean DEFAULT false NOT NULL,
    discarded_at timestamp without time zone,
    discarded_by integer,
    content_locale character varying,
    link_previews jsonb DEFAULT '[]'::jsonb NOT NULL,
    tags character varying[] DEFAULT '{}'::character varying[],
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL,
    poll_template_keys_or_ids jsonb DEFAULT '[]'::jsonb NOT NULL,
    title_placeholder character varying,
    public boolean DEFAULT false NOT NULL,
    recipient_audience character varying
);


--
-- Name: discussion_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.discussion_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: discussion_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.discussion_templates_id_seq OWNED BY public.discussion_templates.id;


--
-- Name: discussions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.discussions (
    id integer NOT NULL,
    group_id integer,
    author_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    title character varying(255),
    last_comment_at timestamp without time zone,
    description text,
    items_count integer DEFAULT 0 NOT NULL,
    closed_at timestamp without time zone,
    private boolean DEFAULT true NOT NULL,
    key character varying(255),
    iframe_src character varying(255),
    last_activity_at timestamp without time zone,
    last_sequence_id integer DEFAULT 0 NOT NULL,
    first_sequence_id integer DEFAULT 0 NOT NULL,
    versions_count integer DEFAULT 0,
    closed_polls_count integer DEFAULT 0 NOT NULL,
    importance integer DEFAULT 0 NOT NULL,
    seen_by_count integer DEFAULT 0 NOT NULL,
    ranges_string character varying,
    description_format character varying(10) DEFAULT 'md'::character varying NOT NULL,
    attachments jsonb DEFAULT '[]'::jsonb NOT NULL,
    info jsonb DEFAULT '{}'::jsonb NOT NULL,
    max_depth integer DEFAULT 2 NOT NULL,
    newest_first boolean DEFAULT false NOT NULL,
    discarded_at timestamp without time zone,
    members_count integer,
    anonymous_polls_count integer DEFAULT 0 NOT NULL,
    content_locale character varying,
    link_previews jsonb DEFAULT '[]'::jsonb NOT NULL,
    pinned_at timestamp without time zone,
    discarded_by integer,
    template boolean DEFAULT false NOT NULL,
    tags character varying[] DEFAULT '{}'::character varying[],
    discussion_template_id integer,
    discussion_template_key character varying,
    closer_id integer
);


--
-- Name: discussions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.discussions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: discussions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.discussions_id_seq OWNED BY public.discussions.id;


--
-- Name: documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.documents (
    id integer NOT NULL,
    model_id integer,
    model_type character varying,
    title character varying,
    url character varying,
    doctype character varying NOT NULL,
    color character varying NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    icon character varying,
    author_id integer NOT NULL,
    web_url character varying,
    thumb_url character varying,
    file_file_name character varying,
    file_content_type character varying,
    group_id integer
);


--
-- Name: documents_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.documents_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: documents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.documents_id_seq OWNED BY public.documents.id;


--
-- Name: events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.events (
    id integer NOT NULL,
    kind character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    eventable_id integer,
    eventable_type character varying(255),
    user_id integer,
    discussion_id integer,
    sequence_id integer,
    announcement boolean DEFAULT false NOT NULL,
    custom_fields jsonb DEFAULT '{}'::jsonb NOT NULL,
    parent_id integer,
    "position" integer DEFAULT 0 NOT NULL,
    child_count integer DEFAULT 0 NOT NULL,
    depth integer DEFAULT 0 NOT NULL,
    pinned boolean DEFAULT false NOT NULL,
    position_key character varying,
    descendant_count integer DEFAULT 0 NOT NULL,
    eventable_version_id integer
);


--
-- Name: events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.events_id_seq OWNED BY public.events.id;


--
-- Name: forward_email_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.forward_email_rules (
    id bigint NOT NULL,
    handle public.citext NOT NULL,
    email character varying
);


--
-- Name: forward_email_rules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.forward_email_rules_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: forward_email_rules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.forward_email_rules_id_seq OWNED BY public.forward_email_rules.id;


--
-- Name: group_identities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.group_identities (
    id integer NOT NULL,
    group_id integer NOT NULL,
    identity_id integer NOT NULL,
    custom_fields jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: group_identities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.group_identities_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: group_identities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.group_identities_id_seq OWNED BY public.group_identities.id;


--
-- Name: group_surveys; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.group_surveys (
    id bigint NOT NULL,
    group_id integer NOT NULL,
    category character varying,
    location character varying,
    size character varying,
    declaration character varying,
    purpose text,
    usage character varying,
    referrer character varying,
    role character varying,
    website character varying,
    misc text,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    desired_feature character varying,
    segment character varying
);


--
-- Name: group_surveys_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.group_surveys_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: group_surveys_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.group_surveys_id_seq OWNED BY public.group_surveys.id;


--
-- Name: groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.groups (
    id integer NOT NULL,
    name character varying(255),
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    parent_id integer,
    description text,
    memberships_count integer DEFAULT 0 NOT NULL,
    archived_at timestamp without time zone,
    discussions_count integer DEFAULT 0 NOT NULL,
    full_name character varying(255),
    parent_members_can_see_discussions boolean DEFAULT false NOT NULL,
    key character varying(255),
    category_id integer,
    handle public.citext,
    theme_id integer,
    is_visible_to_public boolean DEFAULT true NOT NULL,
    is_visible_to_parent_members boolean DEFAULT false NOT NULL,
    discussion_privacy_options character varying DEFAULT 'private_only'::character varying NOT NULL,
    members_can_add_members boolean DEFAULT false NOT NULL,
    membership_granted_upon character varying DEFAULT 'approval'::character varying NOT NULL,
    members_can_edit_discussions boolean DEFAULT true NOT NULL,
    cover_photo_file_name character varying(255),
    cover_photo_content_type character varying(255),
    cover_photo_file_size integer,
    cover_photo_updated_at timestamp without time zone,
    logo_file_name character varying(255),
    logo_content_type character varying(255),
    logo_file_size integer,
    logo_updated_at timestamp without time zone,
    members_can_edit_comments boolean DEFAULT true,
    members_can_raise_motions boolean DEFAULT true NOT NULL,
    members_can_vote boolean DEFAULT true NOT NULL,
    members_can_start_discussions boolean DEFAULT true NOT NULL,
    members_can_create_subgroups boolean DEFAULT false NOT NULL,
    creator_id integer,
    is_referral boolean DEFAULT false NOT NULL,
    cohort_id integer,
    default_group_cover_id integer,
    subscription_id integer,
    invitations_count integer DEFAULT 0 NOT NULL,
    admin_memberships_count integer DEFAULT 0 NOT NULL,
    public_discussions_count integer DEFAULT 0 NOT NULL,
    country character varying,
    region character varying,
    city character varying,
    closed_motions_count integer DEFAULT 0 NOT NULL,
    proposal_outcomes_count integer DEFAULT 0 NOT NULL,
    pending_memberships_count integer DEFAULT 0 NOT NULL,
    recent_activity_count integer DEFAULT 0 NOT NULL,
    closed_polls_count integer DEFAULT 0 NOT NULL,
    polls_count integer DEFAULT 0 NOT NULL,
    subgroups_count integer DEFAULT 0 NOT NULL,
    open_discussions_count integer DEFAULT 0 NOT NULL,
    closed_discussions_count integer DEFAULT 0 NOT NULL,
    token character varying,
    admin_tags character varying,
    members_can_announce boolean DEFAULT true NOT NULL,
    description_format character varying(10) DEFAULT 'md'::character varying NOT NULL,
    attachments jsonb DEFAULT '[]'::jsonb NOT NULL,
    info jsonb DEFAULT '{}'::jsonb NOT NULL,
    new_threads_max_depth integer DEFAULT 3 NOT NULL,
    new_threads_newest_first boolean DEFAULT false NOT NULL,
    admins_can_edit_user_content boolean DEFAULT true NOT NULL,
    listed_in_explore boolean DEFAULT false NOT NULL,
    content_locale character varying,
    members_can_add_guests boolean DEFAULT true NOT NULL,
    members_can_delete_comments boolean DEFAULT true NOT NULL,
    link_previews jsonb DEFAULT '[]'::jsonb NOT NULL,
    discussion_templates_count integer DEFAULT 0 NOT NULL,
    poll_templates_count integer DEFAULT 0 NOT NULL,
    request_to_join_prompt character varying,
    delegates_count integer DEFAULT 0 NOT NULL,
    category character varying,
    can_start_polls_without_discussion boolean DEFAULT false NOT NULL
);


--
-- Name: groups_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.groups_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: groups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.groups_id_seq OWNED BY public.groups.id;


--
-- Name: login_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.login_tokens (
    id integer NOT NULL,
    user_id integer,
    token character varying,
    used boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    redirect character varying,
    code integer NOT NULL,
    is_reactivation boolean DEFAULT false NOT NULL
);


--
-- Name: login_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.login_tokens_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: login_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.login_tokens_id_seq OWNED BY public.login_tokens.id;


--
-- Name: member_email_aliases; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.member_email_aliases (
    id bigint NOT NULL,
    email public.citext NOT NULL,
    require_dkim boolean DEFAULT true NOT NULL,
    require_spf boolean DEFAULT true NOT NULL,
    user_id integer,
    group_id integer NOT NULL,
    author_id integer NOT NULL,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: member_email_aliases_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.member_email_aliases_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: member_email_aliases_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.member_email_aliases_id_seq OWNED BY public.member_email_aliases.id;


--
-- Name: membership_requests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.membership_requests (
    id integer NOT NULL,
    name character varying(255),
    email character varying(255),
    introduction text,
    group_id integer,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    requestor_id integer,
    responder_id integer,
    response character varying(255),
    responded_at timestamp without time zone
);


--
-- Name: membership_requests_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.membership_requests_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: membership_requests_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.membership_requests_id_seq OWNED BY public.membership_requests.id;


--
-- Name: memberships; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.memberships (
    id integer NOT NULL,
    group_id integer,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    inviter_id integer,
    revoked_at timestamp without time zone,
    inbox_position integer DEFAULT 0,
    admin boolean DEFAULT false NOT NULL,
    volume integer,
    experiences jsonb DEFAULT '{}'::jsonb NOT NULL,
    invitation_id integer,
    token character varying,
    accepted_at timestamp without time zone,
    title character varying,
    saml_session_expires_at timestamp without time zone,
    revoker_id integer,
    delegate boolean DEFAULT false NOT NULL
);


--
-- Name: memberships_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.memberships_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: memberships_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.memberships_id_seq OWNED BY public.memberships.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id integer NOT NULL,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    event_id integer,
    viewed boolean DEFAULT false NOT NULL,
    translation_values jsonb DEFAULT '{}'::jsonb NOT NULL,
    url character varying,
    actor_id integer
);


--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.notifications_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;


--
-- Name: oauth_access_grants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.oauth_access_grants (
    id integer NOT NULL,
    resource_owner_id integer NOT NULL,
    application_id integer NOT NULL,
    token character varying NOT NULL,
    expires_in integer NOT NULL,
    redirect_uri text NOT NULL,
    created_at timestamp without time zone NOT NULL,
    revoked_at timestamp without time zone,
    scopes character varying
);


--
-- Name: oauth_access_grants_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.oauth_access_grants_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: oauth_access_grants_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.oauth_access_grants_id_seq OWNED BY public.oauth_access_grants.id;


--
-- Name: oauth_access_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.oauth_access_tokens (
    id integer NOT NULL,
    resource_owner_id integer,
    application_id integer,
    token character varying NOT NULL,
    refresh_token character varying,
    expires_in integer,
    revoked_at timestamp without time zone,
    created_at timestamp without time zone NOT NULL,
    scopes character varying
);


--
-- Name: oauth_access_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.oauth_access_tokens_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: oauth_access_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.oauth_access_tokens_id_seq OWNED BY public.oauth_access_tokens.id;


--
-- Name: oauth_applications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.oauth_applications (
    id integer NOT NULL,
    name character varying NOT NULL,
    uid character varying NOT NULL,
    secret character varying NOT NULL,
    redirect_uri text NOT NULL,
    scopes character varying DEFAULT ''::character varying NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    owner_id integer,
    owner_type character varying,
    logo_file_name character varying,
    logo_content_type character varying,
    logo_file_size integer,
    logo_updated_at timestamp without time zone
);


--
-- Name: oauth_applications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.oauth_applications_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: oauth_applications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.oauth_applications_id_seq OWNED BY public.oauth_applications.id;


--
-- Name: omniauth_identities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.omniauth_identities (
    id integer NOT NULL,
    user_id integer,
    email character varying(255),
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    identity_type character varying(255),
    uid character varying(255),
    name character varying(255),
    access_token character varying DEFAULT ''::character varying,
    logo character varying,
    custom_fields jsonb DEFAULT '{}'::jsonb NOT NULL
);


--
-- Name: omniauth_identities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.omniauth_identities_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: omniauth_identities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.omniauth_identities_id_seq OWNED BY public.omniauth_identities.id;


--
-- Name: outcomes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.outcomes (
    id integer NOT NULL,
    poll_id integer,
    statement text NOT NULL,
    author_id integer NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    latest boolean DEFAULT true NOT NULL,
    poll_option_id integer,
    custom_fields jsonb DEFAULT '{}'::jsonb NOT NULL,
    statement_format character varying(10) DEFAULT 'md'::character varying NOT NULL,
    attachments jsonb DEFAULT '[]'::jsonb NOT NULL,
    versions_count integer DEFAULT 0 NOT NULL,
    review_on date,
    content_locale character varying,
    link_previews jsonb DEFAULT '[]'::jsonb NOT NULL
);


--
-- Name: outcomes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.outcomes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: outcomes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.outcomes_id_seq OWNED BY public.outcomes.id;


--
-- Name: partition_sequences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.partition_sequences (
    key text NOT NULL,
    id integer NOT NULL,
    counter integer DEFAULT 0
);


--
-- Name: pg_search_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pg_search_documents (
    id bigint NOT NULL,
    content text,
    ts_content tsvector,
    author_id bigint,
    group_id bigint,
    discussion_id bigint,
    poll_id bigint,
    searchable_type character varying,
    searchable_id bigint,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL,
    authored_at timestamp(6) without time zone
);


--
-- Name: pg_search_documents_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pg_search_documents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pg_search_documents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.pg_search_documents_id_seq OWNED BY public.pg_search_documents.id;


--
-- Name: poll_options; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.poll_options (
    id integer NOT NULL,
    name character varying NOT NULL,
    poll_id integer,
    priority integer DEFAULT 0 NOT NULL,
    score_counts jsonb DEFAULT '{}'::jsonb NOT NULL,
    voter_scores jsonb DEFAULT '{}'::jsonb NOT NULL,
    total_score integer DEFAULT 0 NOT NULL,
    voter_count integer DEFAULT 0 NOT NULL,
    icon character varying,
    meaning character varying,
    prompt character varying,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    test_operator character varying,
    test_percent integer,
    test_against character varying
);


--
-- Name: poll_options_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.poll_options_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: poll_options_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.poll_options_id_seq OWNED BY public.poll_options.id;


--
-- Name: poll_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.poll_templates (
    id bigint NOT NULL,
    key character varying,
    group_id integer NOT NULL,
    "position" integer DEFAULT 0 NOT NULL,
    author_id integer NOT NULL,
    poll_type character varying NOT NULL,
    process_name character varying,
    process_subtitle character varying,
    title character varying,
    details text,
    details_format character varying(10) DEFAULT 'md'::character varying NOT NULL,
    anonymous boolean DEFAULT false NOT NULL,
    specified_voters_only boolean DEFAULT false NOT NULL,
    notify_on_closing_soon integer DEFAULT 0 NOT NULL,
    content_locale character varying,
    shuffle_options boolean DEFAULT false NOT NULL,
    hide_results integer DEFAULT 0 NOT NULL,
    chart_type character varying,
    min_score integer,
    max_score integer,
    minimum_stance_choices integer,
    maximum_stance_choices integer,
    dots_per_person integer,
    reason_prompt character varying,
    poll_options jsonb DEFAULT '[]'::jsonb NOT NULL,
    stance_reason_required integer DEFAULT 1 NOT NULL,
    limit_reason_length boolean DEFAULT true NOT NULL,
    default_duration_in_days integer DEFAULT 7 NOT NULL,
    agree_target integer,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL,
    meeting_duration integer,
    can_respond_maybe boolean DEFAULT true NOT NULL,
    poll_option_name_format character varying DEFAULT 'plain'::character varying NOT NULL,
    link_previews jsonb DEFAULT '[]'::jsonb NOT NULL,
    atttachments jsonb DEFAULT '[]'::jsonb NOT NULL,
    tags character varying[] DEFAULT '{}'::character varying[],
    discarded_at timestamp(6) without time zone,
    process_introduction character varying,
    process_introduction_format character varying DEFAULT 'md'::character varying NOT NULL,
    title_placeholder character varying,
    outcome_statement character varying,
    outcome_statement_format character varying DEFAULT 'html'::character varying NOT NULL,
    outcome_review_due_in_days integer,
    public boolean DEFAULT false NOT NULL,
    show_none_of_the_above boolean DEFAULT false NOT NULL,
    quorum_pct integer
);


--
-- Name: poll_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.poll_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: poll_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.poll_templates_id_seq OWNED BY public.poll_templates.id;


--
-- Name: polls; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.polls (
    id integer NOT NULL,
    author_id integer NOT NULL,
    title character varying NOT NULL,
    details text,
    closing_at timestamp without time zone,
    closed_at timestamp without time zone,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    discussion_id integer,
    key character varying NOT NULL,
    poll_type character varying NOT NULL,
    stance_data jsonb DEFAULT '{}'::jsonb,
    voters_count integer DEFAULT 0 NOT NULL,
    multiple_choice boolean DEFAULT false NOT NULL,
    custom_fields jsonb DEFAULT '{}'::jsonb NOT NULL,
    stance_counts jsonb DEFAULT '[]'::jsonb NOT NULL,
    group_id integer,
    matrix_counts jsonb DEFAULT '[]'::jsonb NOT NULL,
    undecided_voters_count integer DEFAULT 0 NOT NULL,
    voter_can_add_options boolean DEFAULT false NOT NULL,
    anonymous boolean DEFAULT false NOT NULL,
    versions_count integer DEFAULT 0,
    details_format character varying(10) DEFAULT 'md'::character varying NOT NULL,
    attachments jsonb DEFAULT '[]'::jsonb NOT NULL,
    discarded_at timestamp without time zone,
    discarded_by integer,
    specified_voters_only boolean DEFAULT false NOT NULL,
    notify_on_closing_soon integer DEFAULT 0 NOT NULL,
    content_locale character varying,
    link_previews jsonb DEFAULT '[]'::jsonb NOT NULL,
    shuffle_options boolean DEFAULT false NOT NULL,
    hide_results integer DEFAULT 0 NOT NULL,
    chart_type character varying,
    min_score integer,
    max_score integer,
    minimum_stance_choices integer,
    maximum_stance_choices integer,
    dots_per_person integer,
    process_name character varying,
    template boolean DEFAULT false NOT NULL,
    reason_prompt character varying,
    poll_option_name_format character varying,
    stance_reason_required integer DEFAULT 1 NOT NULL,
    limit_reason_length boolean DEFAULT true NOT NULL,
    default_duration_in_days integer,
    agree_target integer,
    process_subtitle character varying,
    process_url character varying,
    tags character varying[] DEFAULT '{}'::character varying[],
    poll_template_id integer,
    poll_template_key character varying,
    show_none_of_the_above boolean DEFAULT false NOT NULL,
    none_of_the_above_count integer DEFAULT 0 NOT NULL,
    quorum_pct integer
);


--
-- Name: polls_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.polls_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: polls_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.polls_id_seq OWNED BY public.polls.id;


--
-- Name: reactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reactions (
    id integer NOT NULL,
    reactable_id integer,
    user_id integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    reaction character varying DEFAULT '+1'::character varying NOT NULL,
    reactable_type character varying DEFAULT 'Comment'::character varying NOT NULL
);


--
-- Name: reactions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.reactions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: reactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.reactions_id_seq OWNED BY public.reactions.id;


--
-- Name: received_emails; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.received_emails (
    id bigint NOT NULL,
    group_id integer,
    headers public.hstore DEFAULT ''::public.hstore NOT NULL,
    body_text character varying,
    body_html character varying,
    spf_valid boolean DEFAULT false NOT NULL,
    dkim_valid boolean DEFAULT false NOT NULL,
    released boolean DEFAULT false NOT NULL,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: received_emails_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.received_emails_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: received_emails_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.received_emails_id_seq OWNED BY public.received_emails.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: stance_choices; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stance_choices (
    id integer NOT NULL,
    stance_id integer,
    poll_option_id integer,
    score integer DEFAULT 1 NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone
);


--
-- Name: stance_choices_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.stance_choices_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: stance_choices_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.stance_choices_id_seq OWNED BY public.stance_choices.id;


--
-- Name: stance_receipts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stance_receipts (
    id bigint NOT NULL,
    poll_id bigint,
    voter_id bigint,
    inviter_id bigint,
    invited_at timestamp(6) without time zone,
    vote_cast boolean,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: stance_receipts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.stance_receipts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: stance_receipts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.stance_receipts_id_seq OWNED BY public.stance_receipts.id;


--
-- Name: stances; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stances (
    id integer NOT NULL,
    poll_id integer NOT NULL,
    participant_id integer,
    reason character varying,
    latest boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    versions_count integer DEFAULT 0,
    reason_format character varying(10) DEFAULT 'md'::character varying NOT NULL,
    attachments jsonb DEFAULT '[]'::jsonb NOT NULL,
    cast_at timestamp without time zone,
    token character varying,
    revoked_at timestamp without time zone,
    admin boolean DEFAULT false NOT NULL,
    inviter_id integer,
    volume integer DEFAULT 2 NOT NULL,
    accepted_at timestamp without time zone,
    content_locale character varying,
    link_previews jsonb DEFAULT '[]'::jsonb NOT NULL,
    option_scores jsonb DEFAULT '{}'::jsonb NOT NULL,
    revoker_id integer,
    guest boolean DEFAULT false NOT NULL,
    none_of_the_above boolean DEFAULT false NOT NULL
);


--
-- Name: stances_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.stances_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: stances_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.stances_id_seq OWNED BY public.stances.id;


--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.subscriptions (
    id integer NOT NULL,
    expires_at timestamp without time zone,
    chargify_subscription_id integer,
    plan character varying DEFAULT 'free'::character varying,
    payment_method character varying DEFAULT 'none'::character varying NOT NULL,
    owner_id integer,
    max_threads integer,
    max_members integer,
    max_orgs integer,
    state character varying DEFAULT 'active'::character varying NOT NULL,
    members_count integer,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    info jsonb,
    canceled_at timestamp without time zone,
    activated_at timestamp without time zone,
    renews_at timestamp without time zone,
    renewed_at timestamp without time zone,
    allow_subgroups boolean DEFAULT true NOT NULL,
    allow_guests boolean DEFAULT true NOT NULL,
    lead_status character varying
);


--
-- Name: subscriptions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.subscriptions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: subscriptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.subscriptions_id_seq OWNED BY public.subscriptions.id;


--
-- Name: taggings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.taggings (
    id integer NOT NULL,
    tag_id integer NOT NULL,
    taggable_id integer NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    taggable_type character varying NOT NULL
);


--
-- Name: taggings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.taggings_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: taggings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.taggings_id_seq OWNED BY public.taggings.id;


--
-- Name: tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tags (
    id integer NOT NULL,
    group_id integer,
    name public.citext NOT NULL,
    color character varying,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    taggings_count integer DEFAULT 0,
    priority integer DEFAULT 0 NOT NULL,
    org_taggings_count integer DEFAULT 0 NOT NULL
);


--
-- Name: tags_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tags_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tags_id_seq OWNED BY public.tags.id;


--
-- Name: tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tasks (
    id bigint NOT NULL,
    record_type character varying,
    record_id bigint,
    author_id bigint NOT NULL,
    uid integer NOT NULL,
    name character varying NOT NULL,
    done boolean NOT NULL,
    done_at timestamp without time zone,
    due_on date,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL,
    discarded_at timestamp without time zone,
    doer_id integer,
    remind integer,
    remind_at timestamp without time zone
);


--
-- Name: tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tasks_id_seq OWNED BY public.tasks.id;


--
-- Name: tasks_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tasks_users (
    id bigint NOT NULL,
    task_id bigint,
    user_id bigint,
    created_at timestamp(6) without time zone NOT NULL,
    updated_at timestamp(6) without time zone NOT NULL
);


--
-- Name: tasks_users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tasks_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tasks_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tasks_users_id_seq OWNED BY public.tasks_users.id;


--
-- Name: translations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.translations (
    id integer NOT NULL,
    translatable_id integer,
    translatable_type character varying(255),
    fields public.hstore,
    language character varying(255),
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


--
-- Name: translations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.translations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: translations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.translations_id_seq OWNED BY public.translations.id;


--
-- Name: user_deactivation_responses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_deactivation_responses (
    id integer NOT NULL,
    user_id integer,
    body text
);


--
-- Name: user_deactivation_responses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_deactivation_responses_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_deactivation_responses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_deactivation_responses_id_seq OWNED BY public.user_deactivation_responses.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id integer NOT NULL,
    email public.citext,
    encrypted_password character varying(128) DEFAULT ''::character varying,
    reset_password_token character varying(255),
    reset_password_sent_at timestamp without time zone,
    remember_created_at timestamp without time zone,
    sign_in_count integer DEFAULT 0,
    current_sign_in_at timestamp without time zone,
    last_sign_in_at timestamp without time zone,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    name character varying(255),
    deactivated_at timestamp without time zone,
    is_admin boolean DEFAULT false,
    avatar_kind character varying(255) DEFAULT 'initials'::character varying NOT NULL,
    uploaded_avatar_file_name character varying(255),
    uploaded_avatar_content_type character varying(255),
    uploaded_avatar_file_size integer,
    uploaded_avatar_updated_at timestamp without time zone,
    avatar_initials character varying(255),
    username character varying(255),
    email_when_proposal_closing_soon boolean DEFAULT false NOT NULL,
    authentication_token character varying(255),
    unsubscribe_token character varying(255),
    memberships_count integer DEFAULT 0 NOT NULL,
    selected_locale character varying(255),
    time_zone character varying(255),
    key character varying(255),
    detected_locale character varying(255),
    email_catch_up boolean DEFAULT true NOT NULL,
    email_api_key character varying(255),
    email_when_mentioned boolean DEFAULT true NOT NULL,
    email_on_participation boolean DEFAULT false NOT NULL,
    default_membership_volume integer DEFAULT 2 NOT NULL,
    country character varying,
    region character varying,
    city character varying,
    experiences jsonb DEFAULT '{}'::jsonb NOT NULL,
    facebook_community_id integer,
    slack_community_id integer,
    remember_token character varying,
    short_bio character varying DEFAULT ''::character varying NOT NULL,
    email_verified boolean DEFAULT false NOT NULL,
    location character varying DEFAULT ''::character varying NOT NULL,
    last_seen_at timestamp without time zone,
    legal_accepted_at timestamp without time zone,
    email_newsletter boolean DEFAULT false NOT NULL,
    failed_attempts integer DEFAULT 0 NOT NULL,
    unlock_token character varying,
    locked_at timestamp without time zone,
    short_bio_format character varying(10) DEFAULT 'md'::character varying NOT NULL,
    attachments jsonb DEFAULT '[]'::jsonb NOT NULL,
    current_sign_in_ip inet,
    last_sign_in_ip inet,
    secret_token character varying DEFAULT public.gen_random_uuid() NOT NULL,
    content_locale character varying,
    bot boolean DEFAULT false NOT NULL,
    link_previews jsonb DEFAULT '[]'::jsonb NOT NULL,
    email_catch_up_day integer,
    date_time_pref character varying,
    api_key character varying,
    deactivator_id integer,
    autodetect_time_zone boolean DEFAULT true NOT NULL,
    email_sha256 character varying,
    complaints_count integer DEFAULT 0 NOT NULL,
    auto_translate boolean DEFAULT false NOT NULL
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.versions (
    id integer NOT NULL,
    item_type character varying(255) NOT NULL,
    item_id integer NOT NULL,
    event character varying(255) NOT NULL,
    whodunnit integer,
    created_at timestamp without time zone,
    object_changes jsonb
);


--
-- Name: versions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.versions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.versions_id_seq OWNED BY public.versions.id;


--
-- Name: webhooks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.webhooks (
    id bigint NOT NULL,
    group_id integer NOT NULL,
    name character varying NOT NULL,
    url character varying,
    event_kinds jsonb DEFAULT '[]'::jsonb NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    format character varying DEFAULT 'markdown'::character varying,
    include_body boolean DEFAULT false,
    include_subgroups boolean DEFAULT false NOT NULL,
    is_broken boolean DEFAULT false NOT NULL,
    token character varying,
    author_id integer,
    actor_id integer,
    permissions character varying[] DEFAULT '{}'::character varying[] NOT NULL,
    last_used_at timestamp(6) without time zone
);


--
-- Name: webhooks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: webhooks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.webhooks_id_seq OWNED BY public.webhooks.id;


--
-- Name: action_mailbox_inbound_emails id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_mailbox_inbound_emails ALTER COLUMN id SET DEFAULT nextval('public.action_mailbox_inbound_emails_id_seq'::regclass);


--
-- Name: active_admin_comments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.active_admin_comments ALTER COLUMN id SET DEFAULT nextval('public.active_admin_comments_id_seq'::regclass);


--
-- Name: active_storage_attachments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.active_storage_attachments ALTER COLUMN id SET DEFAULT nextval('public.active_storage_attachments_id_seq'::regclass);


--
-- Name: active_storage_blobs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.active_storage_blobs ALTER COLUMN id SET DEFAULT nextval('public.active_storage_blobs_id_seq'::regclass);


--
-- Name: active_storage_variant_records id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.active_storage_variant_records ALTER COLUMN id SET DEFAULT nextval('public.active_storage_variant_records_id_seq'::regclass);


--
-- Name: attachments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attachments ALTER COLUMN id SET DEFAULT nextval('public.attachments_id_seq'::regclass);


--
-- Name: blazer_audits id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blazer_audits ALTER COLUMN id SET DEFAULT nextval('public.blazer_audits_id_seq'::regclass);


--
-- Name: blazer_checks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blazer_checks ALTER COLUMN id SET DEFAULT nextval('public.blazer_checks_id_seq'::regclass);


--
-- Name: blazer_dashboard_queries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blazer_dashboard_queries ALTER COLUMN id SET DEFAULT nextval('public.blazer_dashboard_queries_id_seq'::regclass);


--
-- Name: blazer_dashboards id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blazer_dashboards ALTER COLUMN id SET DEFAULT nextval('public.blazer_dashboards_id_seq'::regclass);


--
-- Name: blazer_queries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blazer_queries ALTER COLUMN id SET DEFAULT nextval('public.blazer_queries_id_seq'::regclass);


--
-- Name: blocked_domains id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocked_domains ALTER COLUMN id SET DEFAULT nextval('public.blocked_domains_id_seq'::regclass);


--
-- Name: chatbots id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chatbots ALTER COLUMN id SET DEFAULT nextval('public.chatbots_id_seq'::regclass);


--
-- Name: cohorts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cohorts ALTER COLUMN id SET DEFAULT nextval('public.cohorts_id_seq'::regclass);


--
-- Name: comments id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments ALTER COLUMN id SET DEFAULT nextval('public.comments_id_seq'::regclass);


--
-- Name: default_group_covers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.default_group_covers ALTER COLUMN id SET DEFAULT nextval('public.default_group_covers_id_seq'::regclass);


--
-- Name: demos id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.demos ALTER COLUMN id SET DEFAULT nextval('public.demos_id_seq'::regclass);


--
-- Name: discussion_readers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.discussion_readers ALTER COLUMN id SET DEFAULT nextval('public.discussion_readers_id_seq'::regclass);


--
-- Name: discussion_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.discussion_templates ALTER COLUMN id SET DEFAULT nextval('public.discussion_templates_id_seq'::regclass);


--
-- Name: discussions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.discussions ALTER COLUMN id SET DEFAULT nextval('public.discussions_id_seq'::regclass);


--
-- Name: documents id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.documents ALTER COLUMN id SET DEFAULT nextval('public.documents_id_seq'::regclass);


--
-- Name: events id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events ALTER COLUMN id SET DEFAULT nextval('public.events_id_seq'::regclass);


--
-- Name: forward_email_rules id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forward_email_rules ALTER COLUMN id SET DEFAULT nextval('public.forward_email_rules_id_seq'::regclass);


--
-- Name: group_identities id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_identities ALTER COLUMN id SET DEFAULT nextval('public.group_identities_id_seq'::regclass);


--
-- Name: group_surveys id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_surveys ALTER COLUMN id SET DEFAULT nextval('public.group_surveys_id_seq'::regclass);


--
-- Name: groups id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.groups ALTER COLUMN id SET DEFAULT nextval('public.groups_id_seq'::regclass);


--
-- Name: login_tokens id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.login_tokens ALTER COLUMN id SET DEFAULT nextval('public.login_tokens_id_seq'::regclass);


--
-- Name: member_email_aliases id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.member_email_aliases ALTER COLUMN id SET DEFAULT nextval('public.member_email_aliases_id_seq'::regclass);


--
-- Name: membership_requests id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.membership_requests ALTER COLUMN id SET DEFAULT nextval('public.membership_requests_id_seq'::regclass);


--
-- Name: memberships id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.memberships ALTER COLUMN id SET DEFAULT nextval('public.memberships_id_seq'::regclass);


--
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);


--
-- Name: oauth_access_grants id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_access_grants ALTER COLUMN id SET DEFAULT nextval('public.oauth_access_grants_id_seq'::regclass);


--
-- Name: oauth_access_tokens id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_access_tokens ALTER COLUMN id SET DEFAULT nextval('public.oauth_access_tokens_id_seq'::regclass);


--
-- Name: oauth_applications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_applications ALTER COLUMN id SET DEFAULT nextval('public.oauth_applications_id_seq'::regclass);


--
-- Name: omniauth_identities id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.omniauth_identities ALTER COLUMN id SET DEFAULT nextval('public.omniauth_identities_id_seq'::regclass);


--
-- Name: outcomes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.outcomes ALTER COLUMN id SET DEFAULT nextval('public.outcomes_id_seq'::regclass);


--
-- Name: pg_search_documents id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pg_search_documents ALTER COLUMN id SET DEFAULT nextval('public.pg_search_documents_id_seq'::regclass);


--
-- Name: poll_options id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.poll_options ALTER COLUMN id SET DEFAULT nextval('public.poll_options_id_seq'::regclass);


--
-- Name: poll_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.poll_templates ALTER COLUMN id SET DEFAULT nextval('public.poll_templates_id_seq'::regclass);


--
-- Name: polls id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.polls ALTER COLUMN id SET DEFAULT nextval('public.polls_id_seq'::regclass);


--
-- Name: reactions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reactions ALTER COLUMN id SET DEFAULT nextval('public.reactions_id_seq'::regclass);


--
-- Name: received_emails id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.received_emails ALTER COLUMN id SET DEFAULT nextval('public.received_emails_id_seq'::regclass);


--
-- Name: stance_choices id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stance_choices ALTER COLUMN id SET DEFAULT nextval('public.stance_choices_id_seq'::regclass);


--
-- Name: stance_receipts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stance_receipts ALTER COLUMN id SET DEFAULT nextval('public.stance_receipts_id_seq'::regclass);


--
-- Name: stances id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stances ALTER COLUMN id SET DEFAULT nextval('public.stances_id_seq'::regclass);


--
-- Name: subscriptions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions ALTER COLUMN id SET DEFAULT nextval('public.subscriptions_id_seq'::regclass);


--
-- Name: taggings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.taggings ALTER COLUMN id SET DEFAULT nextval('public.taggings_id_seq'::regclass);


--
-- Name: tags id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags ALTER COLUMN id SET DEFAULT nextval('public.tags_id_seq'::regclass);


--
-- Name: tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks ALTER COLUMN id SET DEFAULT nextval('public.tasks_id_seq'::regclass);


--
-- Name: tasks_users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks_users ALTER COLUMN id SET DEFAULT nextval('public.tasks_users_id_seq'::regclass);


--
-- Name: translations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.translations ALTER COLUMN id SET DEFAULT nextval('public.translations_id_seq'::regclass);


--
-- Name: user_deactivation_responses id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_deactivation_responses ALTER COLUMN id SET DEFAULT nextval('public.user_deactivation_responses_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: versions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.versions ALTER COLUMN id SET DEFAULT nextval('public.versions_id_seq'::regclass);


--
-- Name: webhooks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhooks ALTER COLUMN id SET DEFAULT nextval('public.webhooks_id_seq'::regclass);


--
-- Name: action_mailbox_inbound_emails action_mailbox_inbound_emails_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.action_mailbox_inbound_emails
    ADD CONSTRAINT action_mailbox_inbound_emails_pkey PRIMARY KEY (id);


--
-- Name: active_admin_comments active_admin_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.active_admin_comments
    ADD CONSTRAINT active_admin_comments_pkey PRIMARY KEY (id);


--
-- Name: active_storage_attachments active_storage_attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.active_storage_attachments
    ADD CONSTRAINT active_storage_attachments_pkey PRIMARY KEY (id);


--
-- Name: active_storage_blobs active_storage_blobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.active_storage_blobs
    ADD CONSTRAINT active_storage_blobs_pkey PRIMARY KEY (id);


--
-- Name: active_storage_variant_records active_storage_variant_records_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.active_storage_variant_records
    ADD CONSTRAINT active_storage_variant_records_pkey PRIMARY KEY (id);


--
-- Name: ar_internal_metadata ar_internal_metadata_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ar_internal_metadata
    ADD CONSTRAINT ar_internal_metadata_pkey PRIMARY KEY (key);


--
-- Name: attachments attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attachments
    ADD CONSTRAINT attachments_pkey PRIMARY KEY (id);


--
-- Name: blazer_audits blazer_audits_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blazer_audits
    ADD CONSTRAINT blazer_audits_pkey PRIMARY KEY (id);


--
-- Name: blazer_checks blazer_checks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blazer_checks
    ADD CONSTRAINT blazer_checks_pkey PRIMARY KEY (id);


--
-- Name: blazer_dashboard_queries blazer_dashboard_queries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blazer_dashboard_queries
    ADD CONSTRAINT blazer_dashboard_queries_pkey PRIMARY KEY (id);


--
-- Name: blazer_dashboards blazer_dashboards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blazer_dashboards
    ADD CONSTRAINT blazer_dashboards_pkey PRIMARY KEY (id);


--
-- Name: blazer_queries blazer_queries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blazer_queries
    ADD CONSTRAINT blazer_queries_pkey PRIMARY KEY (id);


--
-- Name: blocked_domains blocked_domains_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blocked_domains
    ADD CONSTRAINT blocked_domains_pkey PRIMARY KEY (id);


--
-- Name: chatbots chatbots_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chatbots
    ADD CONSTRAINT chatbots_pkey PRIMARY KEY (id);


--
-- Name: cohorts cohorts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cohorts
    ADD CONSTRAINT cohorts_pkey PRIMARY KEY (id);


--
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);


--
-- Name: default_group_covers default_group_covers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.default_group_covers
    ADD CONSTRAINT default_group_covers_pkey PRIMARY KEY (id);


--
-- Name: demos demos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.demos
    ADD CONSTRAINT demos_pkey PRIMARY KEY (id);


--
-- Name: discussion_readers discussion_readers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.discussion_readers
    ADD CONSTRAINT discussion_readers_pkey PRIMARY KEY (id);


--
-- Name: discussion_templates discussion_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.discussion_templates
    ADD CONSTRAINT discussion_templates_pkey PRIMARY KEY (id);


--
-- Name: discussions discussions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.discussions
    ADD CONSTRAINT discussions_pkey PRIMARY KEY (id);


--
-- Name: documents documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- Name: forward_email_rules forward_email_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forward_email_rules
    ADD CONSTRAINT forward_email_rules_pkey PRIMARY KEY (id);


--
-- Name: group_identities group_identities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_identities
    ADD CONSTRAINT group_identities_pkey PRIMARY KEY (id);


--
-- Name: group_surveys group_surveys_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.group_surveys
    ADD CONSTRAINT group_surveys_pkey PRIMARY KEY (id);


--
-- Name: groups groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_pkey PRIMARY KEY (id);


--
-- Name: login_tokens login_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.login_tokens
    ADD CONSTRAINT login_tokens_pkey PRIMARY KEY (id);


--
-- Name: member_email_aliases member_email_aliases_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.member_email_aliases
    ADD CONSTRAINT member_email_aliases_pkey PRIMARY KEY (id);


--
-- Name: membership_requests membership_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.membership_requests
    ADD CONSTRAINT membership_requests_pkey PRIMARY KEY (id);


--
-- Name: memberships memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.memberships
    ADD CONSTRAINT memberships_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: oauth_access_grants oauth_access_grants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_access_grants
    ADD CONSTRAINT oauth_access_grants_pkey PRIMARY KEY (id);


--
-- Name: oauth_access_tokens oauth_access_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_access_tokens
    ADD CONSTRAINT oauth_access_tokens_pkey PRIMARY KEY (id);


--
-- Name: oauth_applications oauth_applications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_applications
    ADD CONSTRAINT oauth_applications_pkey PRIMARY KEY (id);


--
-- Name: omniauth_identities omniauth_identities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.omniauth_identities
    ADD CONSTRAINT omniauth_identities_pkey PRIMARY KEY (id);


--
-- Name: outcomes outcomes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.outcomes
    ADD CONSTRAINT outcomes_pkey PRIMARY KEY (id);


--
-- Name: partition_sequences partition_sequences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.partition_sequences
    ADD CONSTRAINT partition_sequences_pkey PRIMARY KEY (key, id);


--
-- Name: pg_search_documents pg_search_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pg_search_documents
    ADD CONSTRAINT pg_search_documents_pkey PRIMARY KEY (id);


--
-- Name: poll_options poll_options_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.poll_options
    ADD CONSTRAINT poll_options_pkey PRIMARY KEY (id);


--
-- Name: poll_templates poll_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.poll_templates
    ADD CONSTRAINT poll_templates_pkey PRIMARY KEY (id);


--
-- Name: polls polls_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.polls
    ADD CONSTRAINT polls_pkey PRIMARY KEY (id);


--
-- Name: reactions reactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reactions
    ADD CONSTRAINT reactions_pkey PRIMARY KEY (id);


--
-- Name: received_emails received_emails_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.received_emails
    ADD CONSTRAINT received_emails_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: stance_choices stance_choices_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stance_choices
    ADD CONSTRAINT stance_choices_pkey PRIMARY KEY (id);


--
-- Name: stance_receipts stance_receipts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stance_receipts
    ADD CONSTRAINT stance_receipts_pkey PRIMARY KEY (id);


--
-- Name: stances stances_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stances
    ADD CONSTRAINT stances_pkey PRIMARY KEY (id);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (id);


--
-- Name: taggings taggings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.taggings
    ADD CONSTRAINT taggings_pkey PRIMARY KEY (id);


--
-- Name: tags tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- Name: tasks_users tasks_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tasks_users
    ADD CONSTRAINT tasks_users_pkey PRIMARY KEY (id);


--
-- Name: translations translations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.translations
    ADD CONSTRAINT translations_pkey PRIMARY KEY (id);


--
-- Name: user_deactivation_responses user_deactivation_responses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_deactivation_responses
    ADD CONSTRAINT user_deactivation_responses_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: versions versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.versions
    ADD CONSTRAINT versions_pkey PRIMARY KEY (id);


--
-- Name: webhooks webhooks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.webhooks
    ADD CONSTRAINT webhooks_pkey PRIMARY KEY (id);


--
-- Name: active_storage_blobs_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX active_storage_blobs_idx ON public.active_storage_blobs USING btree (id);


--
-- Name: discussion_readers_guests; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX discussion_readers_guests ON public.discussion_readers USING btree (guest) WHERE (guest = true);


--
-- Name: groups_subscription_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX groups_subscription_id_idx ON public.groups USING btree (subscription_id);


--
-- Name: index_action_mailbox_inbound_emails_uniqueness; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_action_mailbox_inbound_emails_uniqueness ON public.action_mailbox_inbound_emails USING btree (message_id, message_checksum);


--
-- Name: index_active_admin_comments_on_author_type_and_author_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_active_admin_comments_on_author_type_and_author_id ON public.active_admin_comments USING btree (author_type, author_id);


--
-- Name: index_active_admin_comments_on_resource_type_and_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_active_admin_comments_on_resource_type_and_resource_id ON public.active_admin_comments USING btree (resource_type, resource_id);


--
-- Name: index_active_storage_attachments_on_blob_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_active_storage_attachments_on_blob_id ON public.active_storage_attachments USING btree (blob_id);


--
-- Name: index_active_storage_attachments_on_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_active_storage_attachments_on_group_id ON public.active_storage_attachments USING btree (group_id);


--
-- Name: index_active_storage_attachments_uniqueness; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_active_storage_attachments_uniqueness ON public.active_storage_attachments USING btree (record_type, record_id, name, blob_id);


--
-- Name: index_active_storage_blobs_on_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_active_storage_blobs_on_key ON public.active_storage_blobs USING btree (key);


--
-- Name: index_active_storage_variant_records_uniqueness; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_active_storage_variant_records_uniqueness ON public.active_storage_variant_records USING btree (blob_id, variation_digest);


--
-- Name: index_attachments_on_attachable_id_and_attachable_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_attachments_on_attachable_id_and_attachable_type ON public.attachments USING btree (attachable_id, attachable_type);


--
-- Name: index_attachments_on_comment_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_attachments_on_comment_id ON public.attachments USING btree (comment_id);


--
-- Name: index_blazer_audits_on_query_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_blazer_audits_on_query_id ON public.blazer_audits USING btree (query_id);


--
-- Name: index_blazer_audits_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_blazer_audits_on_user_id ON public.blazer_audits USING btree (user_id);


--
-- Name: index_blazer_checks_on_creator_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_blazer_checks_on_creator_id ON public.blazer_checks USING btree (creator_id);


--
-- Name: index_blazer_checks_on_query_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_blazer_checks_on_query_id ON public.blazer_checks USING btree (query_id);


--
-- Name: index_blazer_dashboard_queries_on_dashboard_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_blazer_dashboard_queries_on_dashboard_id ON public.blazer_dashboard_queries USING btree (dashboard_id);


--
-- Name: index_blazer_dashboard_queries_on_query_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_blazer_dashboard_queries_on_query_id ON public.blazer_dashboard_queries USING btree (query_id);


--
-- Name: index_blazer_dashboards_on_creator_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_blazer_dashboards_on_creator_id ON public.blazer_dashboards USING btree (creator_id);


--
-- Name: index_blazer_queries_on_creator_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_blazer_queries_on_creator_id ON public.blazer_queries USING btree (creator_id);


--
-- Name: index_blocked_domains_on_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_blocked_domains_on_name ON public.blocked_domains USING btree (name);


--
-- Name: index_chatbots_on_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_chatbots_on_group_id ON public.chatbots USING btree (group_id);


--
-- Name: index_comments_on_discussion_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_comments_on_discussion_id ON public.comments USING btree (discussion_id);


--
-- Name: index_comments_on_parent_type_and_parent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_comments_on_parent_type_and_parent_id ON public.comments USING btree (parent_type, parent_id);


--
-- Name: index_demos_on_author_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_demos_on_author_id ON public.demos USING btree (author_id);


--
-- Name: index_discussion_readers_discussion_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_discussion_readers_discussion_id ON public.discussion_readers USING btree (discussion_id);


--
-- Name: index_discussion_readers_on_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_discussion_readers_on_token ON public.discussion_readers USING btree (token);


--
-- Name: index_discussion_readers_on_user_id_and_discussion_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_discussion_readers_on_user_id_and_discussion_id ON public.discussion_readers USING btree (user_id, discussion_id);


--
-- Name: index_discussion_templates_on_discarded_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_discussion_templates_on_discarded_at ON public.discussion_templates USING btree (discarded_at);


--
-- Name: index_discussions_on_author_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_discussions_on_author_id ON public.discussions USING btree (author_id);


--
-- Name: index_discussions_on_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_discussions_on_created_at ON public.discussions USING btree (created_at);


--
-- Name: index_discussions_on_discarded_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_discussions_on_discarded_at ON public.discussions USING btree (discarded_at) WHERE (discarded_at IS NULL);


--
-- Name: index_discussions_on_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_discussions_on_group_id ON public.discussions USING btree (group_id);


--
-- Name: index_discussions_on_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_discussions_on_key ON public.discussions USING btree (key);


--
-- Name: index_discussions_on_last_activity_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_discussions_on_last_activity_at ON public.discussions USING btree (last_activity_at DESC);


--
-- Name: index_discussions_on_private; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_discussions_on_private ON public.discussions USING btree (private);


--
-- Name: index_discussions_on_tags; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_discussions_on_tags ON public.discussions USING gin (tags);


--
-- Name: index_discussions_on_template; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_discussions_on_template ON public.discussions USING btree (template) WHERE (template IS TRUE);


--
-- Name: index_documents_on_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_documents_on_group_id ON public.documents USING btree (group_id);


--
-- Name: index_documents_on_model_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_documents_on_model_id ON public.documents USING btree (model_id);


--
-- Name: index_documents_on_model_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_documents_on_model_type ON public.documents USING btree (model_type);


--
-- Name: index_events_on_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_events_on_created_at ON public.events USING btree (created_at);


--
-- Name: index_events_on_discussion_id_and_sequence_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_events_on_discussion_id_and_sequence_id ON public.events USING btree (discussion_id, sequence_id);


--
-- Name: index_events_on_eventable_id_and_kind; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_events_on_eventable_id_and_kind ON public.events USING btree (eventable_id, kind);


--
-- Name: index_events_on_eventable_type_and_eventable_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_events_on_eventable_type_and_eventable_id ON public.events USING btree (eventable_type, eventable_id);


--
-- Name: index_events_on_parent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_events_on_parent_id ON public.events USING btree (parent_id);


--
-- Name: index_events_on_parent_id_and_discussion_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_events_on_parent_id_and_discussion_id ON public.events USING btree (parent_id, discussion_id) WHERE (discussion_id IS NOT NULL);


--
-- Name: index_events_on_position_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_events_on_position_key ON public.events USING btree (position_key);


--
-- Name: index_events_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_events_on_user_id ON public.events USING btree (user_id);


--
-- Name: index_group_identities_on_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_group_identities_on_group_id ON public.group_identities USING btree (group_id);


--
-- Name: index_group_identities_on_identity_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_group_identities_on_identity_id ON public.group_identities USING btree (identity_id);


--
-- Name: index_group_surveys_on_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_group_surveys_on_group_id ON public.group_surveys USING btree (group_id);


--
-- Name: index_groups_on_archived_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_groups_on_archived_at ON public.groups USING btree (archived_at) WHERE (archived_at IS NULL);


--
-- Name: index_groups_on_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_groups_on_created_at ON public.groups USING btree (created_at);


--
-- Name: index_groups_on_full_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_groups_on_full_name ON public.groups USING btree (full_name);


--
-- Name: index_groups_on_handle; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_groups_on_handle ON public.groups USING btree (handle);


--
-- Name: index_groups_on_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_groups_on_key ON public.groups USING btree (key);


--
-- Name: index_groups_on_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_groups_on_name ON public.groups USING btree (name);


--
-- Name: index_groups_on_parent_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_groups_on_parent_id ON public.groups USING btree (parent_id);


--
-- Name: index_groups_on_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_groups_on_token ON public.groups USING btree (token);


--
-- Name: index_login_tokens_on_token; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_login_tokens_on_token ON public.login_tokens USING btree (token);


--
-- Name: index_member_email_aliases_on_email_and_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_member_email_aliases_on_email_and_group_id ON public.member_email_aliases USING btree (email, group_id);


--
-- Name: index_membership_requests_on_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_membership_requests_on_group_id ON public.membership_requests USING btree (group_id);


--
-- Name: index_membership_requests_on_requestor_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_membership_requests_on_requestor_id ON public.membership_requests USING btree (requestor_id);


--
-- Name: index_membership_requests_on_responder_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_membership_requests_on_responder_id ON public.membership_requests USING btree (responder_id);


--
-- Name: index_memberships_on_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_memberships_on_created_at ON public.memberships USING btree (created_at);


--
-- Name: index_memberships_on_group_id_and_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_memberships_on_group_id_and_user_id ON public.memberships USING btree (group_id, user_id);


--
-- Name: index_memberships_on_inviter_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_memberships_on_inviter_id ON public.memberships USING btree (inviter_id);


--
-- Name: index_memberships_on_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_memberships_on_token ON public.memberships USING btree (token);


--
-- Name: index_memberships_on_user_id_and_volume; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_memberships_on_user_id_and_volume ON public.memberships USING btree (user_id, volume);


--
-- Name: index_memberships_on_volume; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_memberships_on_volume ON public.memberships USING btree (volume);


--
-- Name: index_notifications_on_event_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_notifications_on_event_id ON public.notifications USING btree (event_id);


--
-- Name: index_notifications_on_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_notifications_on_id ON public.notifications USING btree (id DESC);


--
-- Name: index_notifications_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_notifications_on_user_id ON public.notifications USING btree (user_id);


--
-- Name: index_oauth_access_grants_on_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_oauth_access_grants_on_token ON public.oauth_access_grants USING btree (token);


--
-- Name: index_oauth_access_tokens_on_refresh_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_oauth_access_tokens_on_refresh_token ON public.oauth_access_tokens USING btree (refresh_token);


--
-- Name: index_oauth_access_tokens_on_resource_owner_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_oauth_access_tokens_on_resource_owner_id ON public.oauth_access_tokens USING btree (resource_owner_id);


--
-- Name: index_oauth_access_tokens_on_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_oauth_access_tokens_on_token ON public.oauth_access_tokens USING btree (token);


--
-- Name: index_oauth_applications_on_owner_id_and_owner_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_oauth_applications_on_owner_id_and_owner_type ON public.oauth_applications USING btree (owner_id, owner_type);


--
-- Name: index_oauth_applications_on_uid; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_oauth_applications_on_uid ON public.oauth_applications USING btree (uid);


--
-- Name: index_omniauth_identities_on_identity_type_and_uid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_omniauth_identities_on_identity_type_and_uid ON public.omniauth_identities USING btree (identity_type, uid);


--
-- Name: index_outcomes_on_poll_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_outcomes_on_poll_id ON public.outcomes USING btree (poll_id);


--
-- Name: index_personas_on_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_personas_on_email ON public.omniauth_identities USING btree (email);


--
-- Name: index_personas_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_personas_on_user_id ON public.omniauth_identities USING btree (user_id);


--
-- Name: index_pg_search_documents_on_author_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_pg_search_documents_on_author_id ON public.pg_search_documents USING btree (author_id);


--
-- Name: index_pg_search_documents_on_discussion_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_pg_search_documents_on_discussion_id ON public.pg_search_documents USING btree (discussion_id);


--
-- Name: index_pg_search_documents_on_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_pg_search_documents_on_group_id ON public.pg_search_documents USING btree (group_id);


--
-- Name: index_pg_search_documents_on_poll_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_pg_search_documents_on_poll_id ON public.pg_search_documents USING btree (poll_id);


--
-- Name: index_pg_search_documents_on_searchable; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_pg_search_documents_on_searchable ON public.pg_search_documents USING btree (searchable_type, searchable_id);


--
-- Name: index_poll_options_on_poll_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_poll_options_on_poll_id ON public.poll_options USING btree (poll_id);


--
-- Name: index_poll_templates_on_discarded_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_poll_templates_on_discarded_at ON public.poll_templates USING btree (discarded_at);


--
-- Name: index_polls_on_author_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_polls_on_author_id ON public.polls USING btree (author_id);


--
-- Name: index_polls_on_closed_at_and_closing_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_polls_on_closed_at_and_closing_at ON public.polls USING btree (closed_at, closing_at);


--
-- Name: index_polls_on_closed_at_and_discussion_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_polls_on_closed_at_and_discussion_id ON public.polls USING btree (closed_at, discussion_id);


--
-- Name: index_polls_on_discussion_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_polls_on_discussion_id ON public.polls USING btree (discussion_id);


--
-- Name: index_polls_on_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_polls_on_group_id ON public.polls USING btree (group_id);


--
-- Name: index_polls_on_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_polls_on_key ON public.polls USING btree (key);


--
-- Name: index_polls_on_tags; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_polls_on_tags ON public.polls USING gin (tags);


--
-- Name: index_reactions_on_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_reactions_on_created_at ON public.reactions USING btree (created_at);


--
-- Name: index_reactions_on_reactable_id_and_reactable_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_reactions_on_reactable_id_and_reactable_type ON public.reactions USING btree (reactable_id, reactable_type);


--
-- Name: index_reactions_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_reactions_on_user_id ON public.reactions USING btree (user_id);


--
-- Name: index_stance_choices_on_poll_option_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_stance_choices_on_poll_option_id ON public.stance_choices USING btree (poll_option_id);


--
-- Name: index_stance_choices_on_stance_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_stance_choices_on_stance_id ON public.stance_choices USING btree (stance_id);


--
-- Name: index_stances_on_participant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_stances_on_participant_id ON public.stances USING btree (participant_id);


--
-- Name: index_stances_on_poll_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_stances_on_poll_id ON public.stances USING btree (poll_id);


--
-- Name: index_stances_on_poll_id_and_cast_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_stances_on_poll_id_and_cast_at ON public.stances USING btree (poll_id NULLS FIRST, cast_at NULLS FIRST);


--
-- Name: index_stances_on_poll_id_and_participant_id_and_latest; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_stances_on_poll_id_and_participant_id_and_latest ON public.stances USING btree (poll_id, participant_id, latest) WHERE (latest = true);


--
-- Name: index_stances_on_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_stances_on_token ON public.stances USING btree (token);


--
-- Name: index_subscriptions_on_owner_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_subscriptions_on_owner_id ON public.subscriptions USING btree (owner_id);


--
-- Name: index_subscriptions_on_plan; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_subscriptions_on_plan ON public.subscriptions USING btree (plan);


--
-- Name: index_taggings_on_tag_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taggings_on_tag_id ON public.taggings USING btree (tag_id);


--
-- Name: index_taggings_on_taggable_type_and_taggable_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_taggings_on_taggable_type_and_taggable_id ON public.taggings USING btree (taggable_type, taggable_id);


--
-- Name: index_tags_on_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_tags_on_group_id ON public.tags USING btree (group_id);


--
-- Name: index_tags_on_group_id_and_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_tags_on_group_id_and_name ON public.tags USING btree (group_id, name);


--
-- Name: index_tags_on_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_tags_on_name ON public.tags USING btree (name);


--
-- Name: index_tasks_on_author_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_tasks_on_author_id ON public.tasks USING btree (author_id);


--
-- Name: index_tasks_on_discarded_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_tasks_on_discarded_at ON public.tasks USING btree (discarded_at);


--
-- Name: index_tasks_on_done; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_tasks_on_done ON public.tasks USING btree (done);


--
-- Name: index_tasks_on_due_on; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_tasks_on_due_on ON public.tasks USING btree (due_on);


--
-- Name: index_tasks_on_record_type_and_record_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_tasks_on_record_type_and_record_id ON public.tasks USING btree (record_type, record_id);


--
-- Name: index_tasks_on_remind_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_tasks_on_remind_at ON public.tasks USING btree (remind_at);


--
-- Name: index_tasks_users_on_task_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_tasks_users_on_task_id ON public.tasks_users USING btree (task_id);


--
-- Name: index_tasks_users_on_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_tasks_users_on_user_id ON public.tasks_users USING btree (user_id);


--
-- Name: index_translations_on_language; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_translations_on_language ON public.translations USING btree (language);


--
-- Name: index_translations_on_translatable_type_and_translatable_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_translations_on_translatable_type_and_translatable_id ON public.translations USING btree (translatable_type, translatable_id);


--
-- Name: index_users_on_api_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_api_key ON public.users USING btree (api_key);


--
-- Name: index_users_on_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_users_on_email ON public.users USING btree (email);


--
-- Name: index_users_on_email_verified; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_users_on_email_verified ON public.users USING btree (email_verified);


--
-- Name: index_users_on_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_users_on_key ON public.users USING btree (key);


--
-- Name: index_users_on_reset_password_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_users_on_reset_password_token ON public.users USING btree (reset_password_token);


--
-- Name: index_users_on_unlock_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_users_on_unlock_token ON public.users USING btree (unlock_token);


--
-- Name: index_users_on_unsubscribe_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_users_on_unsubscribe_token ON public.users USING btree (unsubscribe_token);


--
-- Name: index_users_on_username; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX index_users_on_username ON public.users USING btree (username);


--
-- Name: index_versions_on_item_type_and_item_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_versions_on_item_type_and_item_id ON public.versions USING btree (item_type, item_id);


--
-- Name: index_webhooks_on_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX index_webhooks_on_group_id ON public.webhooks USING btree (group_id);


--
-- Name: inviter_id_not_null; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX inviter_id_not_null ON public.discussion_readers USING btree (inviter_id) WHERE (inviter_id IS NOT NULL);


--
-- Name: notifications_user_id_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_user_id_id_idx ON public.notifications USING btree (user_id, id);


--
-- Name: pg_search_documents_authored_at_asc_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX pg_search_documents_authored_at_asc_index ON public.pg_search_documents USING btree (authored_at);


--
-- Name: pg_search_documents_authored_at_desc_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX pg_search_documents_authored_at_desc_index ON public.pg_search_documents USING btree (authored_at DESC);


--
-- Name: pg_search_documents_searchable_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX pg_search_documents_searchable_index ON public.pg_search_documents USING gin (ts_content);


--
-- Name: stances_guests; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX stances_guests ON public.stances USING btree (guest) WHERE (guest = true);


--
-- Name: users_remember_token_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX users_remember_token_idx ON public.users USING btree (remember_token);


--
-- Name: active_storage_variant_records fk_rails_993965df05; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.active_storage_variant_records
    ADD CONSTRAINT fk_rails_993965df05 FOREIGN KEY (blob_id) REFERENCES public.active_storage_blobs(id);


--
-- Name: active_storage_attachments fk_rails_c3b3935057; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.active_storage_attachments
    ADD CONSTRAINT fk_rails_c3b3935057 FOREIGN KEY (blob_id) REFERENCES public.active_storage_blobs(id);


--
-- PostgreSQL database dump complete
--

\unrestrict 2pqiIIKKIQf0yIgtJbuSh2yjI6fskVVQdxPo7hAy63YEIaqGYh56i6D2ZB6YI49

