DROP VIEW public.profiles;

ALTER TABLE auth.users RENAME COLUMN oauth_google_id TO oauth_google_sub;
/* sub and id are conviently exactly the same (tested + fact checked) */

DROP POLICY select_users ON auth.users;
DROP POLICY insert_useres ON auth.users;
DROP POLICY select_users ON auth.users;
DROP POLICY insert_users ON auth.users;
DROP POLICY update_users ON auth.users;
DROP POLICY delete_users ON auth.users;
DROP POLICY select_sessions ON auth.sessions;
DROP POLICY insert_sessions ON auth.sessions;
DROP POLICY update_sessions ON auth.sessions;
DROP POLICY delete_sessions ON auth.sessions;
DROP POLICY select_studysets ON studysets;
DROP POLICY insert_studysets ON studysets;
DROP POLICY update_studysets ON studysets;
DROP POLICY delete_studysets ON studysets;
DROP POLICY select_studyset_progress ON studyset_progress;
DROP POLICY insert_studyset_progress ON studyset_progress;
DROP POLICY update_studyset_progress ON studyset_progress;
DROP POLICY delete_studyset_progress ON studyset_progress;
DROP POLICY select_studyset_settings ON studyset_settings;
DROP POLICY insert_studyset_settings ON studyset_settings;
DROP POLICY update_studyset_settings ON studyset_settings;
DROP POLICY delete_studyset_settings ON studyset_settings;
DROP POLICY select_users_eh_classes ON auth.users;
DROP POLICY select_sessions_eh_classes ON auth.sessions;

ALTER TABLE auth.users DISABLE ROW LEVEL SECURITY;
ALTER TABLE auth.sessions DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.studysets DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.studyset_progress DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.search_queries DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.studyset_settings DISABLE ROW LEVEL SECURITY;

ALTER TYPE auth_type_enum RENAME VALUE 'username_password' TO 'USERNAME_PASSWORD';
ALTER TYPE auth_type_enum RENAME VALUE 'oauth_google' TO 'OAUTH_GOOGLE';
