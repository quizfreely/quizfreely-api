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
-- Name: auth; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA auth;


--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: auth_type_enum; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.auth_type_enum AS ENUM (
    'USERNAME_PASSWORD',
    'OAUTH_GOOGLE'
);


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: sessions; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.sessions (
    id bigint NOT NULL,
    token text DEFAULT encode(public.gen_random_bytes(32), 'base64'::text) NOT NULL,
    user_id uuid NOT NULL,
    expire_at timestamp with time zone DEFAULT (now() + '10 days'::interval)
);


--
-- Name: sessions_id_seq; Type: SEQUENCE; Schema: auth; Owner: -
--

CREATE SEQUENCE auth.sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: auth; Owner: -
--

ALTER SEQUENCE auth.sessions_id_seq OWNED BY auth.sessions.id;


--
-- Name: users; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    username text,
    encrypted_password text,
    display_name text NOT NULL,
    auth_type public.auth_type_enum NOT NULL,
    oauth_google_sub text,
    oauth_google_email text
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: search_queries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.search_queries (
    query text NOT NULL,
    subject text
);


--
-- Name: studyset_progress; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.studyset_progress (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    studyset_id uuid,
    user_id uuid,
    terms jsonb NOT NULL,
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: studyset_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.studyset_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    studyset_id uuid,
    user_id uuid,
    settings jsonb NOT NULL,
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: studysets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.studysets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    title text DEFAULT 'Untitled Studyset'::text NOT NULL,
    private boolean DEFAULT false NOT NULL,
    data jsonb NOT NULL,
    updated_at timestamp with time zone DEFAULT now(),
    terms_count integer,
    featured boolean DEFAULT false,
    tsvector_title tsvector GENERATED ALWAYS AS (to_tsvector('english'::regconfig, title)) STORED
);


--
-- Name: sessions id; Type: DEFAULT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.sessions ALTER COLUMN id SET DEFAULT nextval('auth.sessions_id_seq'::regclass);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: users users_oauth_google_sub_key; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.users
    ADD CONSTRAINT users_oauth_google_sub_key UNIQUE (oauth_google_sub);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: search_queries search_queries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.search_queries
    ADD CONSTRAINT search_queries_pkey PRIMARY KEY (query);


--
-- Name: studyset_progress studyset_progress_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.studyset_progress
    ADD CONSTRAINT studyset_progress_pkey PRIMARY KEY (id);


--
-- Name: studyset_settings studyset_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.studyset_settings
    ADD CONSTRAINT studyset_settings_pkey PRIMARY KEY (id);


--
-- Name: studysets studysets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.studysets
    ADD CONSTRAINT studysets_pkey PRIMARY KEY (id);


--
-- Name: textsearch_title_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX textsearch_title_idx ON public.studysets USING gin (tsvector_title);


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: studyset_progress studyset_progress_studyset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.studyset_progress
    ADD CONSTRAINT studyset_progress_studyset_id_fkey FOREIGN KEY (studyset_id) REFERENCES public.studysets(id) ON DELETE CASCADE;


--
-- Name: studyset_progress studyset_progress_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.studyset_progress
    ADD CONSTRAINT studyset_progress_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: studyset_settings studyset_settings_studyset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.studyset_settings
    ADD CONSTRAINT studyset_settings_studyset_id_fkey FOREIGN KEY (studyset_id) REFERENCES public.studysets(id) ON DELETE CASCADE;


--
-- Name: studyset_settings studyset_settings_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.studyset_settings
    ADD CONSTRAINT studyset_settings_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: studysets studysets_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.studysets
    ADD CONSTRAINT studysets_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE SET NULL;


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20250814012300');
