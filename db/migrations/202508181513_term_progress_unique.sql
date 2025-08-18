-- migrate:up
alter table term_progress
    add constraint term_progress_unique unique (term_id, user_id);

-- migrate:down
