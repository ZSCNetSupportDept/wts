-- ** Database generated with pgModeler (PostgreSQL Database Modeler).
-- ** pgModeler version: 1.2.2
-- ** PostgreSQL version: 17.0
-- ** Project Site: pgmodeler.io
-- ** Model Author: ---
-- object: app | type: ROLE --
-- DROP ROLE IF EXISTS app;
CREATE ROLE app WITH 
	LOGIN
	 PASSWORD 'ZSCNetworkSupport::WTS@2025';
-- ddl-end --
COMMENT ON ROLE app IS E'Web后端系统连接数据库的默认Role';
-- ddl-end --


-- ** Database creation must be performed outside a multi lined SQL file. 
-- ** These commands were put in this file only as a convenience.

-- object: zsc | type: DATABASE --
-- DROP DATABASE IF EXISTS zsc;

-- Prepended SQL commands --
-- 目前RLS我还没配置好，就先关掉了，先把其它部分搞好，正常运行了再来考虑RLS的事情
-- ddl-end --

CREATE DATABASE zsc
	ENCODING = 'UTF8'
	LC_COLLATE = 'C'
	LC_CTYPE = 'C';
-- ddl-end --


SET check_function_bodies = false;
-- ddl-end --

-- object: data | type: SCHEMA --
-- DROP SCHEMA IF EXISTS data CASCADE;
CREATE SCHEMA data;
-- ddl-end --
ALTER SCHEMA data OWNER TO postgres;
-- ddl-end --
COMMENT ON SCHEMA data IS E'存放系统数据与状态等';
-- ddl-end --

-- object: wts | type: SCHEMA --
-- DROP SCHEMA IF EXISTS wts CASCADE;
CREATE SCHEMA wts;
-- ddl-end --
ALTER SCHEMA wts OWNER TO postgres;
-- ddl-end --
COMMENT ON SCHEMA wts IS E'报修系统使用schema';
-- ddl-end --

-- object: scheduler | type: SCHEMA --
-- DROP SCHEMA IF EXISTS scheduler CASCADE;
CREATE SCHEMA scheduler;
-- ddl-end --
ALTER SCHEMA scheduler OWNER TO postgres;
-- ddl-end --
COMMENT ON SCHEMA scheduler IS E'网维成员的排班系统(正在开发中)';
-- ddl-end --

SET search_path TO pg_catalog,public,data,wts,scheduler;
-- ddl-end --

-- object: data.students | type: TABLE --
-- DROP TABLE IF EXISTS data.students CASCADE;
CREATE TABLE data.students (
	sid text NOT NULL,
	name text NOT NULL,
	CONSTRAINT students_pk PRIMARY KEY (sid)
);
-- ddl-end --
COMMENT ON TABLE data.students IS E'学校所有学生的姓名和学号，系统依赖于这个表运行，记录由管理人员负责插入';
-- ddl-end --
COMMENT ON COLUMN data.students.sid IS E'学号';
-- ddl-end --
COMMENT ON COLUMN data.students.name IS E'姓名';
-- ddl-end --
ALTER TABLE data.students OWNER TO postgres;
-- ddl-end --

INSERT INTO data.students (sid, name) VALUES (E'-1', E'(系统操作)');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'-2', E'用户');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd1', E'1栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd2', E'2栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd3', E'3栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd4', E'4栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd5', E'5栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd6', E'6栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd7', E'7栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd8', E'8栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd9', E'9栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd10', E'10栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd11', E'11栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd12', E'12栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd13', E'13栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd14', E'14栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd15', E'15栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd16', E'16栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd17', E'17栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd18', E'18栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd19', E'19栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd20', E'20栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd21', E'21栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gd22', E'22栋工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gdXHA', E'香晖A工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gdXHB', E'香晖B工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gdXHC', E'香晖C工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gdXHD', E'香晖D工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gdZH', E'朝晖工单');
-- ddl-end --
INSERT INTO data.students (sid, name) VALUES (E'gdOther', E'其它片区工单');
-- ddl-end --

-- object: wts.block | type: TYPE --
-- DROP TYPE IF EXISTS wts.block CASCADE;
CREATE TYPE wts.block AS
ENUM ('1','2','3','4','5','6','7','8','9','10','11','12','13','14','15','16','17','18','19','20','21','22','XHA','XHB','XHC','XHD','ZH','other');
-- ddl-end --
ALTER TYPE wts.block OWNER TO postgres;
-- ddl-end --
COMMENT ON TYPE wts.block IS E'宿舍的楼号';
-- ddl-end --

-- object: wts.isp | type: TYPE --
-- DROP TYPE IF EXISTS wts.isp CASCADE;
CREATE TYPE wts.isp AS
ENUM ('telecom','unicom','mobile','others','broadnet');
-- ddl-end --
ALTER TYPE wts.isp OWNER TO postgres;
-- ddl-end --
COMMENT ON TYPE wts.isp IS E'运营商';
-- ddl-end --

-- object: wts.access | type: TYPE --
-- DROP TYPE IF EXISTS wts.access CASCADE;
CREATE TYPE wts.access AS
ENUM ('dev','chief','api','group-leader','formal-member','informal-member','pre-member','user','unregistered');
-- ddl-end --
ALTER TYPE wts.access OWNER TO postgres;
-- ddl-end --

-- object: wts.priority | type: TYPE --
-- DROP TYPE IF EXISTS wts.priority CASCADE;
CREATE TYPE wts.priority AS
ENUM ('highest','assigned','mainline','normal','in-passing','least');
-- ddl-end --
ALTER TYPE wts.priority OWNER TO postgres;
-- ddl-end --
COMMENT ON TYPE wts.priority IS E'工单的优先级';
-- ddl-end --

-- object: wts.status | type: TYPE --
-- DROP TYPE IF EXISTS wts.status CASCADE;
CREATE TYPE wts.status AS
ENUM ('fresh','scheduled','delay','escalated','solved','canceled');
-- ddl-end --
ALTER TYPE wts.status OWNER TO postgres;
-- ddl-end --
COMMENT ON TYPE wts.status IS E'工单的状态';
-- ddl-end --

-- object: wts.users | type: TABLE --
-- DROP TABLE IF EXISTS wts.users CASCADE;
CREATE TABLE wts.users (
	sid text NOT NULL,
	phone text,
	block wts.block,
	room text,
	isp wts.isp,
	account text,
	wx text NOT NULL,
	op boolean NOT NULL DEFAULT false,
	registered_at timestamptz NOT NULL DEFAULT NOW(),
	updated_at timestamptz DEFAULT NOW(),
	CONSTRAINT users_pk PRIMARY KEY (sid),
	CONSTRAINT phone_unique UNIQUE (phone),
	CONSTRAINT check_phone CHECK (phone ~ '^1[3-9]\d{9}$')
);
-- ddl-end --
COMMENT ON TABLE wts.users IS E'报修系统的用户';
-- ddl-end --
COMMENT ON COLUMN wts.users.sid IS E'用户的学号';
-- ddl-end --
COMMENT ON COLUMN wts.users.phone IS E'用于联系用户的手机号';
-- ddl-end --
COMMENT ON COLUMN wts.users.block IS E'楼号';
-- ddl-end --
COMMENT ON COLUMN wts.users.room IS E'房间';
-- ddl-end --
COMMENT ON COLUMN wts.users.isp IS E'运营商';
-- ddl-end --
COMMENT ON COLUMN wts.users.account IS E'宽带的账号';
-- ddl-end --
COMMENT ON COLUMN wts.users.wx IS E'微信(OpenID)';
-- ddl-end --
COMMENT ON COLUMN wts.users.op IS E'是不是网维';
-- ddl-end --
COMMENT ON COLUMN wts.users.registered_at IS E'注册的时间';
-- ddl-end --
COMMENT ON COLUMN wts.users.updated_at IS E'最近个人信息更新的时间';
-- ddl-end --
COMMENT ON CONSTRAINT check_phone ON wts.users IS E'检查手机号';
-- ddl-end --
ALTER TABLE wts.users OWNER TO postgres;
-- ddl-end --

INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'-1', NULL, NULL, NULL, NULL, NULL, E'system', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'-2', NULL, NULL, NULL, NULL, NULL, E'user', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd1', NULL, E'1', E'工单', NULL, NULL, E'1', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd2', NULL, E'2', E'工单', NULL, NULL, E'2', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd3', NULL, E'3', E'工单', NULL, NULL, E'3', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd4', NULL, E'4', E'工单', NULL, NULL, E'4', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd5', NULL, E'5', E'工单', NULL, NULL, E'5', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd6', NULL, E'6', E'工单', NULL, NULL, E'6', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd7', NULL, E'7', E'工单', NULL, NULL, E'7', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd8', NULL, E'8', E'工单', NULL, NULL, E'8', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd9', NULL, E'9', E'工单', NULL, NULL, E'9', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd10', NULL, E'10', E'工单', NULL, NULL, E'10', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd11', NULL, E'11', E'工单', NULL, NULL, E'11', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd12', NULL, E'12', E'工单', NULL, NULL, E'12', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd13', NULL, E'13', E'工单', NULL, NULL, E'13', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd14', NULL, E'14', E'工单', NULL, NULL, E'14', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd15', NULL, E'15', E'工单', NULL, NULL, E'15', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd16', NULL, E'16', E'工单', NULL, NULL, E'16', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd17', NULL, E'17', E'工单', NULL, NULL, E'17', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd18', NULL, E'18', E'工单', NULL, NULL, E'18', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd19', NULL, E'19', E'工单', NULL, NULL, E'19', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd20', NULL, E'20', E'工单', NULL, NULL, E'20', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd21', NULL, E'21', E'工单', NULL, NULL, E'21', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gd22', NULL, E'22', E'工单', NULL, NULL, E'22', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gdXHA', NULL, E'XHA', E'工单', NULL, NULL, E'xha', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gdXHB', NULL, E'XHB', E'工单', NULL, NULL, E'xhb', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gdXHC', NULL, E'XHC', E'工单', NULL, NULL, E'xhc', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gdXHD', NULL, E'XHD', E'工单', NULL, NULL, E'xhd', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gdZH', NULL, E'ZH', E'工单', NULL, NULL, E'zh', E'false', current_timestamp, current_timestamp);
-- ddl-end --
INSERT INTO wts.users (sid, phone, block, room, isp, account, wx, op, registered_at, updated_at) VALUES (E'gdOther', NULL, E'other', E'工单', NULL, NULL, E'other', E'false', current_timestamp, current_timestamp);
-- ddl-end --

-- object: wts.category | type: TYPE --
-- DROP TYPE IF EXISTS wts.category CASCADE;
CREATE TYPE wts.category AS
ENUM ('first-install','low-speed','ip-or-device','client-or-account','others');
-- ddl-end --
ALTER TYPE wts.category OWNER TO postgres;
-- ddl-end --
COMMENT ON TYPE wts.category IS E'故障的类型';
-- ddl-end --

-- object: wts.ticket_traces | type: TABLE --
-- DROP TABLE IF EXISTS wts.ticket_traces CASCADE;
CREATE TABLE wts.ticket_traces (
	opid integer NOT NULL GENERATED ALWAYS AS IDENTITY ,
	tid integer NOT NULL,
	updated_at timestamptz NOT NULL DEFAULT NOW(),
	op text NOT NULL,
	new_status wts.status,
	new_priority wts.priority,
	new_appointment date,
	new_category wts.category,
	remark text NOT NULL,
	CONSTRAINT ticket_traces_pk PRIMARY KEY (opid)
);
-- ddl-end --
COMMENT ON TABLE wts.ticket_traces IS E'工单的情况追踪';
-- ddl-end --
COMMENT ON COLUMN wts.ticket_traces.opid IS E'一个操作的编号，自增主键。';
-- ddl-end --
COMMENT ON COLUMN wts.ticket_traces.tid IS E'工单的编号';
-- ddl-end --
COMMENT ON COLUMN wts.ticket_traces.updated_at IS E'该追踪更新的日期';
-- ddl-end --
COMMENT ON COLUMN wts.ticket_traces.op IS E'进行操作的网维成员';
-- ddl-end --
COMMENT ON COLUMN wts.ticket_traces.new_status IS E'工单的新状态(若有)，没有则NULL';
-- ddl-end --
COMMENT ON COLUMN wts.ticket_traces.new_priority IS E'工单的新优先级，没有则NULL';
-- ddl-end --
COMMENT ON COLUMN wts.ticket_traces.new_appointment IS E'新的预约时间';
-- ddl-end --
COMMENT ON COLUMN wts.ticket_traces.new_category IS E'工单的新类型，没有则NULL';
-- ddl-end --
COMMENT ON COLUMN wts.ticket_traces.remark IS E'本次修改的说明';
-- ddl-end --
ALTER TABLE wts.ticket_traces OWNER TO postgres;
-- ddl-end --

-- object: wts.operators | type: TABLE --
-- DROP TABLE IF EXISTS wts.operators CASCADE;
CREATE TABLE wts.operators (
	wid text NOT NULL,
	sid text NOT NULL,
	access wts.access NOT NULL DEFAULT 'user',
	female boolean NOT NULL,
	CONSTRAINT operators_pk PRIMARY KEY (wid),
	CONSTRAINT sid_unique UNIQUE (sid)
);
-- ddl-end --
COMMENT ON TABLE wts.operators IS E'网维的成员';
-- ddl-end --
COMMENT ON COLUMN wts.operators.wid IS E'工号';
-- ddl-end --
COMMENT ON COLUMN wts.operators.sid IS E'学号';
-- ddl-end --
COMMENT ON COLUMN wts.operators.access IS E'权限';
-- ddl-end --
COMMENT ON COLUMN wts.operators.female IS E'是不是女生';
-- ddl-end --
ALTER TABLE wts.operators OWNER TO postgres;
-- ddl-end --

INSERT INTO wts.operators (wid, sid, access, female) VALUES (E'-1', E'-1', E'user', E'true');
-- ddl-end --
INSERT INTO wts.operators (wid, sid, access, female) VALUES (E'-2', E'-2', E'user', E'false');
-- ddl-end --

-- object: scheduler.weekday | type: TYPE --
-- DROP TYPE IF EXISTS scheduler.weekday CASCADE;
CREATE TYPE scheduler.weekday AS
ENUM ('1','2','3','4','5','6','7');
-- ddl-end --
ALTER TYPE scheduler.weekday OWNER TO postgres;
-- ddl-end --
COMMENT ON TYPE scheduler.weekday IS E'一周的7天';
-- ddl-end --

-- object: scheduler.freeday | type: TABLE --
-- DROP TABLE IF EXISTS scheduler.freeday CASCADE;
CREATE TABLE scheduler.freeday (
	wid text NOT NULL,
	free_at scheduler.weekday NOT NULL,
	CONSTRAINT freeday_pk PRIMARY KEY (wid,free_at)
);
-- ddl-end --
COMMENT ON TABLE scheduler.freeday IS E'每个成员在哪天有空的记录';
-- ddl-end --
COMMENT ON COLUMN scheduler.freeday.wid IS E'工号';
-- ddl-end --
COMMENT ON COLUMN scheduler.freeday.free_at IS E'在哪天有空';
-- ddl-end --
ALTER TABLE scheduler.freeday OWNER TO postgres;
-- ddl-end --

-- object: wts.tickets | type: TABLE --
-- DROP TABLE IF EXISTS wts.tickets CASCADE;
CREATE TABLE wts.tickets (
	tid integer NOT NULL GENERATED ALWAYS AS IDENTITY ,
	issuer text NOT NULL,
	submitted_at timestamptz NOT NULL DEFAULT NOW(),
	category wts.category NOT NULL DEFAULT 'others',
	description text NOT NULL,
	occur_at timestamptz,
	notes text,
	appointed_at date,
	priority wts.priority NOT NULL DEFAULT 'mainline',
	status wts.status NOT NULL DEFAULT 'fresh',
	last_updated_at timestamptz DEFAULT NOW(),
	CONSTRAINT tickets_pk PRIMARY KEY (tid),
	CONSTRAINT occur_at_check CHECK (occur_at <= NOW())
);
-- ddl-end --
COMMENT ON TABLE wts.tickets IS E'工单';
-- ddl-end --
COMMENT ON COLUMN wts.tickets.tid IS E'工单的编号，自增主键';
-- ddl-end --
COMMENT ON COLUMN wts.tickets.issuer IS E'报修人（学号）';
-- ddl-end --
COMMENT ON COLUMN wts.tickets.submitted_at IS E'工单提交时间';
-- ddl-end --
COMMENT ON COLUMN wts.tickets.category IS E'故障的类型';
-- ddl-end --
COMMENT ON COLUMN wts.tickets.description IS E'错误描述，用户所填的..';
-- ddl-end --
COMMENT ON COLUMN wts.tickets.occur_at IS E'网络故障的出现时间（大概）';
-- ddl-end --
COMMENT ON COLUMN wts.tickets.notes IS E'备注，用户所填的';
-- ddl-end --
COMMENT ON COLUMN wts.tickets.appointed_at IS E'预约上门的日期，冗余';
-- ddl-end --
COMMENT ON COLUMN wts.tickets.priority IS E'优先级,冗余';
-- ddl-end --
COMMENT ON COLUMN wts.tickets.status IS E'工单目前的状态,冗余';
-- ddl-end --
COMMENT ON COLUMN wts.tickets.last_updated_at IS E'工单最后更新的时间，冗余';
-- ddl-end --
COMMENT ON CONSTRAINT occur_at_check ON wts.tickets IS E'故障肯定是在之前发生的';
-- ddl-end --
ALTER TABLE wts.tickets OWNER TO postgres;
-- ddl-end --

-- object: wts.sync_ticket | type: FUNCTION --
-- DROP FUNCTION IF EXISTS wts.sync_ticket() CASCADE;
CREATE OR REPLACE FUNCTION wts.sync_ticket ()
	RETURNS trigger
	LANGUAGE plpgsql
	VOLATILE 
	CALLED ON NULL INPUT
	SECURITY DEFINER
	PARALLEL UNSAFE
	COST 100
	SET search_path = pg_catalog, wts
	AS 
$function$
--使用COALESCE()虽然更简洁，但是会有性能问题，所以这版本我改掉了
BEGIN
  UPDATE wts.tickets AS t
  SET
    status = CASE
               WHEN NEW.new_status IS NOT NULL THEN NEW.new_status
               ELSE t.status
             END,
    priority = CASE
                 WHEN NEW.new_priority IS NOT NULL THEN NEW.new_priority
                 ELSE t.priority
               END,
	category = CASE
                 WHEN NEW.new_category IS NOT NULL THEN NEW.new_category
                 ELSE t.category
               END,
    appointed_at = CASE
                     WHEN NEW.new_appointment IS NOT NULL THEN NEW.new_appointment
                     ELSE t.appointed_at
                   END,
    last_updated_at = GREATEST(t.last_updated_at, NEW.updated_at)
  WHERE t.tid = NEW.tid
    AND (
      (NEW.new_status IS NOT NULL AND NEW.new_status IS DISTINCT FROM t.status)
      OR (NEW.new_priority IS NOT NULL AND NEW.new_priority IS DISTINCT FROM t.priority)
      OR (NEW.new_appointment IS NOT NULL AND NEW.new_appointment IS DISTINCT FROM t.appointed_at)
      OR (NEW.updated_at IS NOT NULL AND (t.last_updated_at IS NULL OR NEW.updated_at > t.last_updated_at))
    );

  RETURN NEW;
END;
$function$;
-- ddl-end --
ALTER FUNCTION wts.sync_ticket() OWNER TO postgres;
-- ddl-end --

-- object: sync_on_insert | type: TRIGGER --
-- DROP TRIGGER IF EXISTS sync_on_insert ON wts.ticket_traces CASCADE;
CREATE OR REPLACE TRIGGER sync_on_insert
	AFTER INSERT 
	ON wts.ticket_traces
	FOR EACH ROW
	EXECUTE PROCEDURE wts.sync_ticket();
-- ddl-end --
COMMENT ON TRIGGER sync_on_insert ON wts.ticket_traces IS E'同步数据至冗余';
-- ddl-end --

-- object: wts.v_users | type: VIEW --
-- DROP VIEW IF EXISTS wts.v_users CASCADE;
CREATE OR REPLACE VIEW wts.v_users
AS 
SELECT
s.sid, s.name,
u.phone, u.block, u.room,
u.isp, u.account, u.op,
u.wx, o.access
from wts.users u
LEFT OUTER JOIN data.students s ON u.sid = s.sid
LEFT OUTER JOIN wts.operators o ON o.sid = u.sid;
-- ddl-end --
ALTER VIEW wts.v_users OWNER TO postgres;
-- ddl-end --

-- object: wts.v_tickets | type: VIEW --
-- DROP VIEW IF EXISTS wts.v_tickets CASCADE;
CREATE OR REPLACE VIEW wts.v_tickets
AS 
SELECT t.*,u.name,u.block,u.room,u.isp,u.account,u.phone
FROM wts.tickets t JOIN wts.v_users u ON t.issuer = u.sid;
-- ddl-end --
ALTER VIEW wts.v_tickets OWNER TO postgres;
-- ddl-end --
COMMENT ON VIEW wts.v_tickets IS E'工单的视图';
-- ddl-end --

-- object: idx_ticket_trace_tid_and_updated_at | type: INDEX --
-- DROP INDEX IF EXISTS wts.idx_ticket_trace_tid_and_updated_at CASCADE;
CREATE INDEX idx_ticket_trace_tid_and_updated_at ON wts.ticket_traces
USING btree
(
	tid,
	updated_at DESC NULLS LAST
);
-- ddl-end --
COMMENT ON INDEX wts.idx_ticket_trace_tid_and_updated_at IS E'所指工单和更新时间的复合索引，因为我们常用WHERE tid=a ORDER BY updated_at DESC，这样的查询';
-- ddl-end --

-- object: wts.v_operators | type: VIEW --
-- DROP VIEW IF EXISTS wts.v_operators CASCADE;
CREATE OR REPLACE VIEW wts.v_operators
AS 
SELECT o.wid, u.name, o.access, o.female FROM wts.v_users u RIGHT OUTER JOIN wts.operators o ON u.sid = o.sid;
-- ddl-end --
ALTER VIEW wts.v_operators OWNER TO postgres;
-- ddl-end --
COMMENT ON VIEW wts.v_operators IS E'网维的成员';
-- ddl-end --

-- object: wts.auto_init_trace | type: FUNCTION --
-- DROP FUNCTION IF EXISTS wts.auto_init_trace() CASCADE;
CREATE OR REPLACE FUNCTION wts.auto_init_trace ()
	RETURNS trigger
	LANGUAGE plpgsql
	VOLATILE 
	CALLED ON NULL INPUT
	SECURITY DEFINER
	PARALLEL UNSAFE
	COST 100
	SET search_path = pg_catalog, wts
	AS 
$function$
BEGIN
    INSERT INTO wts.ticket_traces (tid, updated_at, op, new_status, new_appointment, new_priority,new_category, remark)
    VALUES (	NEW.tid, 
		NEW.submitted_at, 
		'-2', 
		COALESCE(NEW.status,'fresh'), 
		COALESCE(NEW.appointed_at,NULL),
		COALESCE(NEW.priority,'mainline'), 
		COALESCE(NEW.category,'others'),
		'提交了报修');
    RETURN NEW;
END;
$function$;
-- ddl-end --
ALTER FUNCTION wts.auto_init_trace() OWNER TO postgres;
-- ddl-end --
COMMENT ON FUNCTION wts.auto_init_trace() IS E'自动创建最初的追踪';
-- ddl-end --

-- object: auto_init_trace | type: TRIGGER --
-- DROP TRIGGER IF EXISTS auto_init_trace ON wts.tickets CASCADE;
CREATE OR REPLACE TRIGGER auto_init_trace
	AFTER INSERT 
	ON wts.tickets
	FOR EACH ROW
	EXECUTE PROCEDURE wts.auto_init_trace();
-- ddl-end --
COMMENT ON TRIGGER auto_init_trace ON wts.tickets IS E'自动提交初始的trace';
-- ddl-end --

-- object: idx_tickets_issuer | type: INDEX --
-- DROP INDEX IF EXISTS wts.idx_tickets_issuer CASCADE;
CREATE INDEX idx_tickets_issuer ON wts.tickets
USING btree
(
	issuer,
	submitted_at DESC NULLS LAST
);
-- ddl-end --
COMMENT ON INDEX wts.idx_tickets_issuer IS E'加快用户报修的查询';
-- ddl-end --

-- object: idx_tickets_sub_and_pr | type: INDEX --
-- DROP INDEX IF EXISTS wts.idx_tickets_sub_and_pr CASCADE;
CREATE INDEX idx_tickets_sub_and_pr ON wts.tickets
USING btree
(
	priority ASC NULLS LAST,
	submitted_at DESC NULLS LAST
)
WHERE (status IN ('fresh','scheduled','delay','escalated'));
-- ddl-end --

-- object: idx_wx_unique | type: INDEX --
-- DROP INDEX IF EXISTS wts.idx_wx_unique CASCADE;
CREATE UNIQUE INDEX idx_wx_unique ON wts.users
USING btree
(
	wx
)
INCLUDE (op);
-- ddl-end --
COMMENT ON INDEX wts.idx_wx_unique IS E'索引，一个微信只能绑定一个用户';
-- ddl-end --

-- object: wts.auto_update_user_audit | type: FUNCTION --
-- DROP FUNCTION IF EXISTS wts.auto_update_user_audit() CASCADE;
CREATE OR REPLACE FUNCTION wts.auto_update_user_audit ()
	RETURNS trigger
	LANGUAGE plpgsql
	VOLATILE 
	CALLED ON NULL INPUT
	SECURITY INVOKER
	PARALLEL UNSAFE
	COST 1
	AS 
$function$
BEGIN
  NEW.updated_at := NOW();RETURN NEW;
END 
$function$;
-- ddl-end --
ALTER FUNCTION wts.auto_update_user_audit() OWNER TO postgres;
-- ddl-end --
COMMENT ON FUNCTION wts.auto_update_user_audit() IS E'自动更新user.updated_at';
-- ddl-end --

-- object: "autoUpdate" | type: TRIGGER --
-- DROP TRIGGER IF EXISTS "autoUpdate" ON wts.users CASCADE;
CREATE OR REPLACE TRIGGER "autoUpdate"
	BEFORE UPDATE
	ON wts.users
	FOR EACH ROW
	EXECUTE PROCEDURE wts.auto_update_user_audit();
-- ddl-end --
COMMENT ON TRIGGER "autoUpdate" ON wts.users IS E'自动更新～';
-- ddl-end --

-- object: idx_tickets_appointed_at | type: INDEX --
-- DROP INDEX IF EXISTS wts.idx_tickets_appointed_at CASCADE;
CREATE INDEX idx_tickets_appointed_at ON wts.tickets
USING btree
(
	appointed_at DESC NULLS LAST
)
WHERE (status IN ('scheduled'));
-- ddl-end --
COMMENT ON INDEX wts.idx_tickets_appointed_at IS E'加快预约查询';
-- ddl-end --

-- object: wts.is_op | type: FUNCTION --
-- DROP FUNCTION IF EXISTS wts.is_op(wts.access) CASCADE;
CREATE OR REPLACE FUNCTION wts.is_op (a wts.access)
	RETURNS bool
	LANGUAGE sql
	IMMUTABLE 
	CALLED ON NULL INPUT
	SECURITY INVOKER
	PARALLEL UNSAFE
	COST 1
	AS 
$function$
SELECT a IN ('informal-member','formal-member','group-leader','chief','dev','api')
$function$;
-- ddl-end --
ALTER FUNCTION wts.is_op(wts.access) OWNER TO postgres;
-- ddl-end --

-- object: wts.is_mgr | type: FUNCTION --
-- DROP FUNCTION IF EXISTS wts.is_mgr(wts.access) CASCADE;
CREATE OR REPLACE FUNCTION wts.is_mgr (a wts.access)
	RETURNS bool
	LANGUAGE sql
	IMMUTABLE 
	CALLED ON NULL INPUT
	SECURITY INVOKER
	PARALLEL UNSAFE
	COST 1
	AS 
$function$
 SELECT a IN ('group-leader','chief','dev','api')
$function$;
-- ddl-end --
ALTER FUNCTION wts.is_mgr(wts.access) OWNER TO postgres;
-- ddl-end --

-- object: wts.am_i_op | type: FUNCTION --
-- DROP FUNCTION IF EXISTS wts.am_i_op() CASCADE;
CREATE OR REPLACE FUNCTION wts.am_i_op ()
	RETURNS bool
	LANGUAGE sql
	STABLE 
	CALLED ON NULL INPUT
	SECURITY INVOKER
	PARALLEL UNSAFE
	COST 1
	AS 
$function$
SELECT u.op
FROM wts.users u
WHERE u.wx = current_setting('wts.wx', true)
LIMIT 1
$function$;
-- ddl-end --
ALTER FUNCTION wts.am_i_op() OWNER TO postgres;
-- ddl-end --
COMMENT ON FUNCTION wts.am_i_op() IS E'当前用户是不是网维的成员';
-- ddl-end --

-- object: wts.am_i_mgr | type: FUNCTION --
-- DROP FUNCTION IF EXISTS wts.am_i_mgr() CASCADE;
CREATE OR REPLACE FUNCTION wts.am_i_mgr ()
	RETURNS bool
	LANGUAGE sql
	STABLE 
	CALLED ON NULL INPUT
	SECURITY INVOKER
	PARALLEL UNSAFE
	COST 1
	AS 
$function$
SELECT COALESCE(wts.is_mgr(o.access),false)
FROM wts.users u
LEFT OUTER JOIN wts.operators o ON u.sid = o.sid
WHERE u.wx = current_setting('wts.wx', true)
LIMIT 1
$function$;
-- ddl-end --
ALTER FUNCTION wts.am_i_mgr() OWNER TO postgres;
-- ddl-end --
COMMENT ON FUNCTION wts.am_i_mgr() IS E'当前用户是不是网维管理层';
-- ddl-end --

-- object: only_view_self | type: POLICY --
-- DROP POLICY IF EXISTS only_view_self ON data.students CASCADE;
CREATE POLICY only_view_self ON data.students
	AS RESTRICTIVE
	FOR SELECT
	TO app
	USING (  EXISTS (
    SELECT 1
    FROM wts.users me
    WHERE me.wx = current_setting('wts.wx', true)
      AND me.sid = students.sid
  )
  OR wts.am_i_op());
-- ddl-end --

-- object: read_only_self | type: POLICY --
-- DROP POLICY IF EXISTS read_only_self ON wts.users CASCADE;
CREATE POLICY read_only_self ON wts.users
	AS RESTRICTIVE
	FOR SELECT
	TO app
	USING (  wx = current_setting('wts.wx', true)
  OR wts.am_i_op());
-- ddl-end --

-- object: update_only_self | type: POLICY --
-- DROP POLICY IF EXISTS update_only_self ON wts.users CASCADE;
CREATE POLICY update_only_self ON wts.users
	AS RESTRICTIVE
	FOR UPDATE
	TO app
	USING (wx = current_setting('wts.wx', true))
	WITH CHECK (wx = current_setting('wx', true));
-- ddl-end --

-- object: insert_only_self | type: POLICY --
-- DROP POLICY IF EXISTS insert_only_self ON wts.users CASCADE;
CREATE POLICY insert_only_self ON wts.users
	AS RESTRICTIVE
	FOR INSERT
	TO app
	WITH CHECK (  wx = current_setting('wts.wx', true)
  AND op = false);
-- ddl-end --

-- object: read_only_self | type: POLICY --
-- DROP POLICY IF EXISTS read_only_self ON wts.tickets CASCADE;
CREATE POLICY read_only_self ON wts.tickets
	AS RESTRICTIVE
	FOR SELECT
	TO app
	USING (  EXISTS (
    SELECT 1 FROM wts.users me
    WHERE me.wx = current_setting('wts.wx', true)
      AND me.sid = tickets.issuer
  )
  OR wts.am_i_op());
-- ddl-end --

-- object: user_new_ticket | type: POLICY --
-- DROP POLICY IF EXISTS user_new_ticket ON wts.tickets CASCADE;
CREATE POLICY user_new_ticket ON wts.tickets
	AS RESTRICTIVE
	FOR INSERT
	TO app
	WITH CHECK (  EXISTS (
    SELECT 1 FROM wts.users me
    WHERE me.wx = current_setting('wts.wx', true)
      AND me.sid = tickets.issuer
  )
  OR wts.am_i_mgr());
-- ddl-end --

-- object: idx_tickets_status | type: INDEX --
-- DROP INDEX IF EXISTS wts.idx_tickets_status CASCADE;
CREATE INDEX idx_tickets_status ON wts.tickets
USING btree
(
	status ASC NULLS LAST
)
INCLUDE (priority);
-- ddl-end --
COMMENT ON INDEX wts.idx_tickets_status IS E'加快按状态的查询';
-- ddl-end --

-- object: wts.sync_access | type: FUNCTION --
-- DROP FUNCTION IF EXISTS wts.sync_access() CASCADE;
CREATE OR REPLACE FUNCTION wts.sync_access ()
	RETURNS trigger
	LANGUAGE plpgsql
	VOLATILE 
	CALLED ON NULL INPUT
	SECURITY INVOKER
	PARALLEL UNSAFE
	COST 100
	AS 
$function$
DECLARE
	target text;
	new_access wts.access;
	new_op boolean;
BEGIN
	IF(TG_OP = 'DELETE') THEN
		target := OLD.sid;
		new_op := false;
	ELSE -- INSERT or UPDATE , SELECT is safe to ingore
		target := NEW.sid;
		new_access := NEW.access;
		IF wts.is_op(new_access) THEN new_op := true;
		ELSE new_op := false;
		END IF;
	END IF;
	UPDATE wts.users SET op = new_op WHERE sid = target;
	IF (TG_OP = 'DELETE') THEN RETURN OLD;
	ELSE RETURN NEW;
	END IF;
END;
$function$;
-- ddl-end --
ALTER FUNCTION wts.sync_access() OWNER TO postgres;
-- ddl-end --
COMMENT ON FUNCTION wts.sync_access() IS E'自动同步网维成员的权限，不应该直接修改users.op';
-- ddl-end --

-- object: auto_sync_access | type: TRIGGER --
-- DROP TRIGGER IF EXISTS auto_sync_access ON wts.operators CASCADE;
CREATE OR REPLACE TRIGGER auto_sync_access
	AFTER INSERT OR DELETE OR UPDATE
	ON wts.operators
	FOR EACH ROW
	EXECUTE PROCEDURE wts.sync_access();
-- ddl-end --
COMMENT ON TRIGGER auto_sync_access ON wts.operators IS E'自动处理users.op';
-- ddl-end --

-- object: wts.v_active_tickets | type: VIEW --
-- DROP VIEW IF EXISTS wts.v_active_tickets CASCADE;
CREATE OR REPLACE VIEW wts.v_active_tickets
AS 
SELECT * FROM wts.v_tickets WHERE status <> 'solved' AND status <> 'canceled';
-- ddl-end --
ALTER VIEW wts.v_active_tickets OWNER TO postgres;
-- ddl-end --
COMMENT ON VIEW wts.v_active_tickets IS E'活跃的工单';
-- ddl-end --

-- object: sid_fk | type: CONSTRAINT --
-- ALTER TABLE wts.users DROP CONSTRAINT IF EXISTS sid_fk CASCADE;
ALTER TABLE wts.users ADD CONSTRAINT sid_fk FOREIGN KEY (sid)
REFERENCES data.students (sid) MATCH SIMPLE
ON DELETE RESTRICT ON UPDATE CASCADE;
-- ddl-end --
COMMENT ON CONSTRAINT sid_fk ON wts.users IS E'需要在data.students中有记录';
-- ddl-end --


-- object: tid_fk | type: CONSTRAINT --
-- ALTER TABLE wts.ticket_traces DROP CONSTRAINT IF EXISTS tid_fk CASCADE;
ALTER TABLE wts.ticket_traces ADD CONSTRAINT tid_fk FOREIGN KEY (tid)
REFERENCES wts.tickets (tid) MATCH SIMPLE
ON DELETE CASCADE ON UPDATE CASCADE;
-- ddl-end --
COMMENT ON CONSTRAINT tid_fk ON wts.ticket_traces IS E'操作对应的工单对象';
-- ddl-end --


-- object: wid_fk | type: CONSTRAINT --
-- ALTER TABLE wts.ticket_traces DROP CONSTRAINT IF EXISTS wid_fk CASCADE;
ALTER TABLE wts.ticket_traces ADD CONSTRAINT wid_fk FOREIGN KEY (op)
REFERENCES wts.operators (wid) MATCH SIMPLE
ON DELETE RESTRICT ON UPDATE CASCADE;
-- ddl-end --

-- object: sid_fk | type: CONSTRAINT --
-- ALTER TABLE wts.operators DROP CONSTRAINT IF EXISTS sid_fk CASCADE;
ALTER TABLE wts.operators ADD CONSTRAINT sid_fk FOREIGN KEY (sid)
REFERENCES wts.users (sid) MATCH SIMPLE
ON DELETE RESTRICT ON UPDATE CASCADE;
-- ddl-end --
COMMENT ON CONSTRAINT sid_fk ON wts.operators IS E'网维成员先必须是一个用户';
-- ddl-end --


-- object: wid_fk | type: CONSTRAINT --
-- ALTER TABLE scheduler.freeday DROP CONSTRAINT IF EXISTS wid_fk CASCADE;
ALTER TABLE scheduler.freeday ADD CONSTRAINT wid_fk FOREIGN KEY (wid)
REFERENCES wts.operators (wid) MATCH SIMPLE
ON DELETE CASCADE ON UPDATE CASCADE;
-- ddl-end --
COMMENT ON CONSTRAINT wid_fk ON scheduler.freeday IS E'空闲表的wid是网维成员的工号';
-- ddl-end --


-- object: issuer_fk | type: CONSTRAINT --
-- ALTER TABLE wts.tickets DROP CONSTRAINT IF EXISTS issuer_fk CASCADE;
ALTER TABLE wts.tickets ADD CONSTRAINT issuer_fk FOREIGN KEY (issuer)
REFERENCES wts.users (sid) MATCH SIMPLE
ON DELETE RESTRICT ON UPDATE CASCADE;
-- ddl-end --
COMMENT ON CONSTRAINT issuer_fk ON wts.tickets IS E'提交工单的必须是报修系统中存在并绑定的用户';
-- ddl-end --


-- object: "grant_U_0bd9136eec" | type: PERMISSION --
GRANT USAGE
   ON SCHEMA wts
   TO app;

-- ddl-end --


-- object: "grant_U_a6c147604b" | type: PERMISSION --
GRANT USAGE
   ON SCHEMA data
   TO app;

-- ddl-end --


-- object: grant_r_c0d53cb645 | type: PERMISSION --
GRANT SELECT
   ON TABLE data.students
   TO app;

-- ddl-end --


-- object: "grant_X_4bbb359a9b" | type: PERMISSION --
GRANT EXECUTE
   ON FUNCTION wts.auto_init_trace()
   TO app;

-- ddl-end --


-- object: "grant_X_67c340d93a" | type: PERMISSION --
GRANT EXECUTE
   ON FUNCTION wts.auto_update_user_audit()
   TO app;

-- ddl-end --


-- object: "grant_X_02215ce77d" | type: PERMISSION --
GRANT EXECUTE
   ON FUNCTION wts.sync_ticket()
   TO app;

-- ddl-end --


-- object: grant_rawd_1299ae3dc2 | type: PERMISSION --
GRANT SELECT,INSERT,UPDATE,DELETE
   ON TABLE wts.operators
   TO app;

-- ddl-end --


-- object: grant_ra_195e979fed | type: PERMISSION --
GRANT SELECT,INSERT
   ON TABLE wts.ticket_traces
   TO app;

-- ddl-end --


-- object: grant_raw_fad2b38fb7 | type: PERMISSION --
GRANT SELECT,INSERT,UPDATE
   ON TABLE wts.tickets
   TO app;

-- ddl-end --


-- object: grant_raw_c847b1a95f | type: PERMISSION --
GRANT SELECT,INSERT,UPDATE
   ON TABLE wts.users
   TO app;

-- ddl-end --


-- object: grant_r_0862033d5a | type: PERMISSION --
GRANT SELECT
   ON TABLE wts.v_active_tickets
   TO app;

-- ddl-end --


-- object: grant_r_362020f11a | type: PERMISSION --
GRANT SELECT
   ON TABLE wts.v_operators
   TO app;

-- ddl-end --


-- object: grant_r_3921ad4145 | type: PERMISSION --
GRANT SELECT
   ON TABLE wts.v_users
   TO app;

-- ddl-end --


-- object: "revoke_CU_cd8e46e7b6" | type: PERMISSION --
REVOKE CREATE,USAGE
   ON SCHEMA public
   FROM PUBLIC;

-- ddl-end --


-- object: "revoke_CU_1c2277113b" | type: PERMISSION --
REVOKE CREATE,USAGE
   ON SCHEMA data
   FROM PUBLIC;

-- ddl-end --


-- object: "grant_U_e297b68524" | type: PERMISSION --
GRANT USAGE
   ON SCHEMA scheduler
   TO app;

-- ddl-end --


-- object: "revoke_CU_50cca1c3c5" | type: PERMISSION --
REVOKE CREATE,USAGE
   ON SCHEMA scheduler
   FROM PUBLIC;

-- ddl-end --


-- object: "revoke_CU_d0d1ae81e5" | type: PERMISSION --
REVOKE CREATE,USAGE
   ON SCHEMA wts
   FROM PUBLIC;

-- ddl-end --


-- object: grant_c_86df58fb73 | type: PERMISSION --
GRANT CONNECT
   ON DATABASE zsc
   TO app;

-- ddl-end --



