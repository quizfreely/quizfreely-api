-- migrate:up
create table terms (
    id uuid primary key default gen_random_uuid(),
    term text,
    def text,
    studyset_id uuid not null references studysets (id) on delete cascade
);

create table term_progress (
    id uuid primary key default gen_random_uuid(),
    term_id uuid not null references terms (id) on delete cascade,
    user_id uuid not null references auth.users (id) on delete cascade,
    term_first_reviewed_at timestamptz,
    term_last_reviewed_at timestamptz,
    term_review_count int,
    def_first_reviewed_at timestamptz,
    def_last_reviewed_at timestamptz,
    def_review_count int,
    term_leitner_box smallint,
    def_leitner_box smallint
);
-- migrate:down
