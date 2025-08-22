-- migrate:up
alter table term_confusion_pairs
    add constraint confusion_pairs_unique unique (user_id, term_id, confused_term_id, answered_with);

-- migrate:down
