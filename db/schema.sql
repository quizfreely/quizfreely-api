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
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

-- *not* creating schema, since initdb creates it


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
-- Name: answer_with_enum; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.answer_with_enum AS ENUM (
    'TERM',
    'DEF'
);


--
-- Name: auth_type_enum; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.auth_type_enum AS ENUM (
    'USERNAME_PASSWORD',
    'OAUTH_GOOGLE'
);


--
-- Name: submission_action_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.submission_action_type AS ENUM (
    'submit',
    'unsubmit',
    'add_grade',
    'update_grade',
    'remove_grade',
    'add_attachment',
    'update_attachment',
    'remove_attachment'
);


--
-- Name: delete_expired_sessions(); Type: PROCEDURE; Schema: auth; Owner: -
--

CREATE PROCEDURE auth.delete_expired_sessions()
    LANGUAGE sql
    AS $$
delete from auth.sessions where expire_at < (select now())
$$;


--
-- Name: verify_session(text); Type: FUNCTION; Schema: auth; Owner: -
--

CREATE FUNCTION auth.verify_session(session_token text) RETURNS TABLE(user_id uuid)
    LANGUAGE sql
    AS $_$
select user_id from auth.sessions where token = $1 and expire_at > (select now())
$_$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: sessions; Type: TABLE; Schema: auth; Owner: -
--

CREATE TABLE auth.sessions (
    token text DEFAULT encode(public.gen_random_bytes(32), 'base64'::text) NOT NULL,
    user_id uuid NOT NULL,
    expire_at timestamp with time zone DEFAULT (now() + '10 days'::interval)
);


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
-- Name: practice_tests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.practice_tests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    "timestamp" timestamp with time zone DEFAULT now() NOT NULL,
    user_id uuid NOT NULL,
    studyset_id uuid NOT NULL,
    questions_correct smallint,
    questions_total smallint,
    questions jsonb
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
-- Name: studysets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.studysets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    title text NOT NULL,
    private boolean NOT NULL,
    updated_at timestamp with time zone DEFAULT now(),
    terms_count integer,
    featured boolean DEFAULT false,
    tsvector_title tsvector GENERATED ALWAYS AS (to_tsvector('english'::regconfig, title)) STORED
);


--
-- Name: term_confusion_pairs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.term_confusion_pairs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    term_id uuid NOT NULL,
    confused_term_id uuid NOT NULL,
    answered_with public.answer_with_enum NOT NULL,
    confused_count integer,
    last_confused_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: term_progress; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.term_progress (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    term_id uuid NOT NULL,
    user_id uuid NOT NULL,
    term_first_reviewed_at timestamp with time zone,
    term_last_reviewed_at timestamp with time zone,
    term_review_count integer,
    def_first_reviewed_at timestamp with time zone,
    def_last_reviewed_at timestamp with time zone,
    def_review_count integer,
    term_leitner_system_box smallint,
    def_leitner_system_box smallint,
    term_correct_count integer DEFAULT 0 NOT NULL,
    term_incorrect_count integer DEFAULT 0 NOT NULL,
    def_correct_count integer DEFAULT 0 NOT NULL,
    def_incorrect_count integer DEFAULT 0 NOT NULL
);


--
-- Name: terms; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.terms (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    term text,
    def text,
    studyset_id uuid NOT NULL,
    sort_order integer NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (token);


--
-- Name: users users_oauth_google_id_key; Type: CONSTRAINT; Schema: auth; Owner: -
--

ALTER TABLE ONLY auth.users
    ADD CONSTRAINT users_oauth_google_id_key UNIQUE (oauth_google_sub);


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
-- Name: term_confusion_pairs confusion_pairs_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_confusion_pairs
    ADD CONSTRAINT confusion_pairs_unique UNIQUE (user_id, term_id, confused_term_id);


--
-- Name: practice_tests practice_tests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_tests
    ADD CONSTRAINT practice_tests_pkey PRIMARY KEY (id);


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
-- Name: studysets studysets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.studysets
    ADD CONSTRAINT studysets_pkey PRIMARY KEY (id);


--
-- Name: term_confusion_pairs term_confusion_pairs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_confusion_pairs
    ADD CONSTRAINT term_confusion_pairs_pkey PRIMARY KEY (id);


--
-- Name: term_progress term_progress_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_progress
    ADD CONSTRAINT term_progress_pkey PRIMARY KEY (id);


--
-- Name: term_progress term_progress_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_progress
    ADD CONSTRAINT term_progress_unique UNIQUE (term_id, user_id);


--
-- Name: terms terms_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.terms
    ADD CONSTRAINT terms_pkey PRIMARY KEY (id);


--
-- Name: textsearch_title_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX textsearch_title_idx ON public.studysets USING gin (tsvector_title);


--
-- Name: practice_tests practice_tests_studyset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_tests
    ADD CONSTRAINT practice_tests_studyset_id_fkey FOREIGN KEY (studyset_id) REFERENCES public.studysets(id) ON DELETE CASCADE;


--
-- Name: practice_tests practice_tests_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.practice_tests
    ADD CONSTRAINT practice_tests_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: studysets studysets_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.studysets
    ADD CONSTRAINT studysets_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE SET NULL;


--
-- Name: term_confusion_pairs term_confusion_pairs_confused_term_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_confusion_pairs
    ADD CONSTRAINT term_confusion_pairs_confused_term_id_fkey FOREIGN KEY (confused_term_id) REFERENCES public.terms(id) ON DELETE CASCADE;


--
-- Name: term_confusion_pairs term_confusion_pairs_term_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_confusion_pairs
    ADD CONSTRAINT term_confusion_pairs_term_id_fkey FOREIGN KEY (term_id) REFERENCES public.terms(id) ON DELETE CASCADE;


--
-- Name: term_confusion_pairs term_confusion_pairs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_confusion_pairs
    ADD CONSTRAINT term_confusion_pairs_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: term_progress term_progress_term_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_progress
    ADD CONSTRAINT term_progress_term_id_fkey FOREIGN KEY (term_id) REFERENCES public.terms(id) ON DELETE CASCADE;


--
-- Name: term_progress term_progress_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.term_progress
    ADD CONSTRAINT term_progress_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;


--
-- Name: terms terms_studyset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.terms
    ADD CONSTRAINT terms_studyset_id_fkey FOREIGN KEY (studyset_id) REFERENCES public.studysets(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('202508140123'),
    ('202508141431'),
    ('202508181513'),
    ('202508191404'),
    ('202508201847'),
    ('202508202155'),
    ('202508211445');
