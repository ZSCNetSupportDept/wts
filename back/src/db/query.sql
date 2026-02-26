--注意，这里的SQL查询就基本相当于数据库的API，非必要不改动，除非你想大量重构已有的代码 --



--用户管理--

-- name: GetNameBySID :one
SELECT name FROM data.students
WHERE sid = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO wts.users (
  sid, phone, block, room, isp, account, wx
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetUserBySID :one
SELECT * FROM wts.v_users
WHERE sid = $1
LIMIT 1;

-- name: GetUserByWX :one
SELECT * FROM wts.v_users
WHERE wx = $1
LIMIT 1;

-- name: UpdateUser :one
UPDATE wts.users
SET 
  phone = $2,
  block = $3,
  room = $4,
  isp = $5,
  account = $6
WHERE sid = $1
RETURNING *;

-- name: FilterUsers :many
SELECT *
FROM wts.v_users u
WHERE 
    u.name LIKE COALESCE(sqlc.narg('name'), '%')
AND u.phone = COALESCE(sqlc.narg('phone'), u.phone)
AND u.block = COALESCE(sqlc.narg('block'), u.block)
AND u.room = COALESCE(sqlc.narg('room'), u.room)
AND u.isp = COALESCE(sqlc.narg('isp'), u.isp)
AND u.account = COALESCE(sqlc.narg('account'), u.account)
ORDER BY u.sid;


--工单管理--

-- name: CreateTicket :one
INSERT INTO wts.tickets (
  issuer, submitted_at, occur_at, description, appointed_at, notes, priority, category, status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetTicket :one
SELECT * FROM wts.v_tickets
WHERE tid = $1 LIMIT 1;

-- name: ListActiveTickets :many
SELECT * FROM wts.v_active_tickets
ORDER BY priority DESC, submitted_at DESC; --先按优先级，相同的话再按提交时间

-- name: ListActiveTicketsByBlocks :many
SELECT * FROM wts.v_active_tickets
WHERE block = ANY(@blocks::wts.block[])
ORDER BY priority DESC, submitted_at DESC;

-- name: ListTicketsByIssuer :many
SELECT * FROM wts.v_tickets
WHERE issuer = $1
ORDER BY submitted_at DESC;

-- name: ListTicketsByStatus :many
SELECT * FROM wts.v_tickets
WHERE status = $1
ORDER BY submitted_at DESC;

-- name: FilterTickets :many
SELECT *
FROM wts.v_tickets t
WHERE 
    (sqlc.narg('blocks')::wts.block[] IS NULL OR t.block = ANY(sqlc.narg('blocks')::wts.block[]))
AND t.issuer = COALESCE(sqlc.narg('issuer'), t.issuer)
AND (sqlc.narg('category')::wts.category[] IS NULL OR t.category = ANY(sqlc.narg('category')::wts.category[]))
AND (sqlc.narg('isp')::wts.isp[] IS NULL OR t.isp = ANY(sqlc.narg('isp')::wts.isp[]))
AND t.submitted_at >= COALESCE(sqlc.narg('newerThan'), '1970-01-01'::timestamptz)
AND t.submitted_at <= COALESCE(sqlc.narg('olderThan'), NOW()::timestamptz)
AND (sqlc.narg('status')::wts.status[] IS NULL OR t.status = ANY(sqlc.narg('status')::wts.status[]))
ORDER BY t.priority ASC;

-- name: FilterActiveTickets :many
SELECT *
FROM wts.v_active_tickets t
WHERE 
    (sqlc.narg('blocks')::wts.block[] IS NULL OR t.block = ANY(sqlc.narg('blocks')::wts.block[]))
AND t.issuer = COALESCE(sqlc.narg('issuer'), t.issuer)
AND (sqlc.narg('category')::wts.category[] IS NULL OR t.category = ANY(sqlc.narg('category')::wts.category[]))
AND (sqlc.narg('isp')::wts.isp[] IS NULL OR t.isp = ANY(sqlc.narg('isp')::wts.isp[]))
AND t.submitted_at >= COALESCE(sqlc.narg('newerThan'), '1970-01-01'::timestamptz)
AND t.submitted_at <= COALESCE(sqlc.narg('olderThan'), NOW()::timestamptz)
AND (sqlc.narg('status')::wts.status[] IS NULL OR t.status = ANY(sqlc.narg('status')::wts.status[]))
ORDER BY t.priority ASC;


--traces管理 --

-- name: CreateTicketTrace :one
INSERT INTO wts.ticket_traces (
  tid, updated_at, op, new_status, new_priority, new_appointment, new_category, remark
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListTracesByTicket :many
SELECT t.*, o.name
FROM wts.ticket_traces t
LEFT JOIN wts.v_operators o ON o.wid = t.op
WHERE t.tid = $1
ORDER BY t.updated_at DESC;

--网维成员管理--

-- name: GetStaffByWid :one
SELECT * FROM wts.operators
WHERE wid = $1
LIMIT 1;

-- name: GetStaffBySid :one
SELECT * FROM wts.operators
WHERE sid = $1
LIMIT 1;

--数据分析--

-- name: GetActiveTicketCountByBlock :many
SELECT block, COUNT(*) AS total
FROM wts.v_active_tickets
WHERE block IS NOT NULL
GROUP BY block
ORDER BY total DESC;