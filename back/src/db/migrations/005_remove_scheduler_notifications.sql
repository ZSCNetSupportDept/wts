-- Remove only the scheduler's WeChat notification feature.
BEGIN;
DROP FUNCTION IF EXISTS scheduler.notification_template();
DROP TABLE IF EXISTS scheduler.notification_outbox;
DELETE FROM data.kvstore WHERE key = 'WX_NOTIFY_SCHEDULER_TEMPLATE_ID';
COMMIT;
