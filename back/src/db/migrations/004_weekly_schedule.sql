BEGIN;
CREATE TABLE IF NOT EXISTS scheduler.weekly_templates (
 effective_date date PRIMARY KEY,
 weekday integer NOT NULL CHECK(weekday BETWEEN 1 AND 7),
 leader_wid text NOT NULL REFERENCES wts.operators(wid),
 published boolean NOT NULL,
 revision integer NOT NULL,
 members jsonb NOT NULL CHECK(jsonb_typeof(members)='array'),
 updated_by text NOT NULL,
 updated_at timestamptz NOT NULL DEFAULT NOW(),
 CHECK(weekday=extract(isodow FROM effective_date)::integer)
);
ALTER TABLE scheduler.duty_days ADD COLUMN IF NOT EXISTS weekly_source date;
ALTER TABLE scheduler.duty_days ADD COLUMN IF NOT EXISTS weekly_revision integer NOT NULL DEFAULT 0;
ALTER TABLE scheduler.duty_days ADD COLUMN IF NOT EXISTS weekly_exception boolean NOT NULL DEFAULT false;
GRANT SELECT,INSERT,UPDATE ON scheduler.weekly_templates TO app;
COMMIT;
