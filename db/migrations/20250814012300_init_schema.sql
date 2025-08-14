create extension if not exists pgcrypto;
create extension if not exists pg_trgm;

create schema auth;

grant usage on schema auth to quizfreely_api;

create type auth_type_enum as enum (
    'USERNAME_PASSWORD',
    'OAUTH_GOOGLE'
);
create table auth.users (
  id uuid primary key default gen_random_uuid(),
  username text,
  encrypted_password text,
  display_name text not null,
  auth_type auth_type_enum not null,
  oauth_google_sub text,
  oauth_google_email text,
  unique (username),
  unique (oauth_google_id)
);

grant select on auth.users to quizfreely_api;
grant insert on auth.users to quizfreely_api;
grant update on auth.users to quizfreely_api;
grant delete on auth.users to quizfreely_api;

create table auth.sessions (
  id BIGSERIAL PRIMARY KEY,
  token text not null default encode(gen_random_bytes(32), 'base64'),
  user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  expire_at timestamptz default now() + '10 days'::interval
);

grant select on auth.sessions to quizfreely_api;
grant insert on auth.sessions to quizfreely_api;
grant delete on auth.sessions to quizfreely_api;
grant usage, select on auth.sessions_id_seq to quizfreely_api;

create table public.studysets (
  id uuid primary key default gen_random_uuid(),
  user_id uuid references auth.users (id) on delete set null,
  title text not null default 'Untitled Studyset',
  private boolean not null default false,
  data jsonb not null,
  updated_at timestamptz default now(),
  terms_count int,
  featured boolean default false,
  tsvector_title tsvector generated always as (to_tsvector('english', title)) stored
);

create index textsearch_title_idx on public.studysets using GIN (tsvector_title);

grant select on public.studysets to quizfreely_api;
grant insert on public.studysets to quizfreely_api;
grant update on public.studysets to quizfreely_api;
grant delete on public.studysets to quizfreely_api;

create table public.studyset_progress (
  id uuid primary key default gen_random_uuid(),
  studyset_id uuid references public.studysets (id) on delete cascade,
  user_id uuid references auth.users (id) on delete cascade,
  terms jsonb not null,
  updated_at timestamptz default now()
);

grant select on public.studyset_progress to quizfreely_api;
grant insert on public.studyset_progress to quizfreely_api;
grant update on public.studyset_progress to quizfreely_api;
grant delete on public.studyset_progress to quizfreely_api;

create table public.search_queries (
  query text primary key,
  subject text
);

grant select on public.search_queries to quizfreely_api;
grant insert on public.search_queries to quizfreely_api;
grant update on public.search_queries to quizfreely_api;
grant delete on public.search_queries to quizfreely_api;

create table public.studyset_settings (
  id uuid primary key default gen_random_uuid(),
  studyset_id uuid references public.studysets (id) on delete cascade,
  user_id uuid references auth.users (id) on delete cascade,
  settings jsonb not null,
  updated_at timestamptz default now()
);

grant select on public.studyset_settings to quizfreely_api;
grant insert on public.studyset_settings to quizfreely_api;
grant update on public.studyset_settings to quizfreely_api;
grant delete on public.studyset_settings to quizfreely_api;
