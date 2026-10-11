-- Run against the existing application database as its owner before deploying.
BEGIN;
CREATE TABLE scheduler.duty_days (
    duty_date date PRIMARY KEY,
    leader_wid text NOT NULL REFERENCES wts.operators(wid) ON DELETE RESTRICT,
    published boolean NOT NULL DEFAULT false,
    revision integer NOT NULL DEFAULT 1 CHECK (revision > 0),
    updated_by text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE scheduler.duty_members (
    duty_date date NOT NULL REFERENCES scheduler.duty_days(duty_date) ON DELETE CASCADE,
    wid text NOT NULL REFERENCES wts.operators(wid) ON DELETE RESTRICT,
    zone text NOT NULL DEFAULT '' CHECK (zone IN ('','FX','BM','DM','QT','XHAB','XHCD','ZH','other')),
    area_leader boolean NOT NULL DEFAULT false,
    extra boolean NOT NULL DEFAULT false,
    PRIMARY KEY (duty_date, wid)
);
CREATE INDEX duty_members_wid_date ON scheduler.duty_members(wid, duty_date);
CREATE TABLE scheduler.availability_exceptions (
    wid text NOT NULL REFERENCES wts.operators(wid) ON DELETE CASCADE,
    duty_date date NOT NULL,
    available boolean NOT NULL,
    PRIMARY KEY (wid, duty_date)
);
CREATE TABLE scheduler.audit_log (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    duty_date date,
    actor_wid text NOT NULL,
    action text NOT NULL,
    reason text NOT NULL DEFAULT '',
    before_state jsonb,
    after_state jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE scheduler.swap_requests (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    requester_wid text NOT NULL REFERENCES wts.operators(wid),
    target_wid text REFERENCES wts.operators(wid),
    date_a date NOT NULL REFERENCES scheduler.duty_days(duty_date),
    date_b date NOT NULL REFERENCES scheduler.duty_days(duty_date),
    revision_a integer NOT NULL,
    revision_b integer NOT NULL,
    approval_a text REFERENCES wts.operators(wid),
    approval_b text REFERENCES wts.operators(wid),
    status text NOT NULL DEFAULT 'open' CHECK (status IN ('open','pending','approved','rejected','canceled','stale')),
    reason text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (date_a <> date_b),
    CHECK (target_wid IS NULL OR requester_wid <> target_wid),
    CHECK (status NOT IN ('pending','approved') OR target_wid IS NOT NULL)
);
CREATE INDEX swap_requests_pending ON scheduler.swap_requests(date_a, date_b) WHERE status IN ('open','pending');
GRANT USAGE ON SCHEMA scheduler TO app;
GRANT SELECT, INSERT, UPDATE, DELETE ON scheduler.duty_days, scheduler.duty_members,
    scheduler.availability_exceptions, scheduler.freeday, scheduler.swap_requests TO app;
GRANT SELECT, INSERT ON scheduler.audit_log TO app;
GRANT USAGE, SELECT ON SEQUENCE scheduler.audit_log_id_seq TO app;
GRANT USAGE, SELECT ON SEQUENCE scheduler.swap_requests_id_seq TO app;
COMMIT;
