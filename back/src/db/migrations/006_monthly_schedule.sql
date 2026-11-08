BEGIN;
ALTER TABLE scheduler.weekly_templates ADD COLUMN IF NOT EXISTS ends_on date;
UPDATE scheduler.weekly_templates SET ends_on=(effective_date+INTERVAL '1 month')::date WHERE ends_on IS NULL;
ALTER TABLE scheduler.weekly_templates ALTER COLUMN ends_on SET NOT NULL;
-- Keep dated exceptions and explicit schedules. Retire old automatic occurrences beyond the window.
UPDATE scheduler.swap_requests r SET status='stale'
WHERE r.status IN ('open','pending') AND EXISTS (
 SELECT 1 FROM scheduler.duty_days d JOIN scheduler.weekly_templates t ON t.effective_date=d.weekly_source
 WHERE d.duty_date>=t.ends_on AND NOT d.weekly_exception AND d.duty_date>=CURRENT_DATE
 AND d.duty_date IN (r.date_a,r.date_b)
);
UPDATE scheduler.duty_days d SET published=false,revision=d.revision+1
FROM scheduler.weekly_templates t
WHERE d.weekly_source=t.effective_date AND d.duty_date>=t.ends_on AND NOT d.weekly_exception
 AND d.duty_date>=CURRENT_DATE AND d.published;
DELETE FROM scheduler.duty_days d USING scheduler.weekly_templates t
WHERE d.weekly_source=t.effective_date AND d.duty_date>=t.ends_on AND NOT d.weekly_exception
 AND d.duty_date>=CURRENT_DATE
 AND NOT EXISTS(SELECT 1 FROM scheduler.swap_requests r WHERE r.date_a=d.duty_date OR r.date_b=d.duty_date);
COMMIT;
