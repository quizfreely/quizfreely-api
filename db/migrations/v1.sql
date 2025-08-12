ALTER TABLE auth.users RENAME COLUMN oauth_google_id TO oauth_google_sub;
/* sub and id are conviently exactly the same (tested + fact checked) */
