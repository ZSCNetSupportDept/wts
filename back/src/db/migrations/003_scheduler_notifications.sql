BEGIN;
CREATE TABLE IF NOT EXISTS scheduler.notification_outbox (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_key text NOT NULL,
    recipient_wid text NOT NULL REFERENCES wts.operators(wid),
    kind text NOT NULL CHECK (kind IN ('swap', 'extra')),
    date_a date NOT NULL,
    date_b date,
    member text NOT NULL,
    remark text NOT NULL,
    swap_id bigint REFERENCES scheduler.swap_requests(id),
    attempts integer NOT NULL DEFAULT 0,
    next_attempt_at timestamptz NOT NULL DEFAULT NOW(),
    delivered_at timestamptz,
    last_error text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE(event_key, recipient_wid)
);
CREATE INDEX IF NOT EXISTS notification_outbox_pending ON scheduler.notification_outbox(next_attempt_at) WHERE delivered_at IS NULL;
GRANT SELECT,INSERT,UPDATE ON scheduler.notification_outbox TO app;
GRANT USAGE,SELECT ON SEQUENCE scheduler.notification_outbox_id_seq TO app;
CREATE OR REPLACE FUNCTION scheduler.notification_template()
RETURNS text LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $$
    SELECT value FROM data.kvstore WHERE key='WX_NOTIFY_SCHEDULER_TEMPLATE_ID'
$$;
REVOKE ALL ON FUNCTION scheduler.notification_template() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION scheduler.notification_template() TO app;
COMMIT;
