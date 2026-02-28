package daemon

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"zsxyww.com/wts/model"
	"zsxyww.com/wts/model/sqlc"
	"zsxyww.com/wts/server"

	"math/rand"
)

// 在每天值班结束的时候，自动取消（改日修）预约在今天但是状态今天没有更新的工单
func scheduledAutoCancel() {
	go func() {
		var first = true
		var duration time.Duration
		var jobID int
		for {
			jobID = rand.Int()
			//暂时在每晚的9点执行逻辑
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, now.Location())

			// 如果程序启动时已经过了9点，那么就立即执行
			if (!now.Before(next)) && (first == true) {
				first = false
				goto do
			}

			// 如果当前时间已经过了9点，比如程序执行完上一次了，就设置为明天9点的定时
			if !now.Before(next) {
				next = next.AddDate(0, 0, 1)
			}
			first = false //这里防止程序在9点前启动的时候会跑两次逻辑...

			duration = next.Sub(now)
			slog.Info("下一次取消程序安排上了", "duration", duration, "nextTime", next)
			time.Sleep(duration)

			// 醒来后执行取消工单的操作
		do:
			if err := doCancelJob(jobID); err != nil {
				slog.Error("自动预约处理程序执行失败", "error", err, "ID", jobID)
			} else {
				slog.Info("自动取消程序执行完毕", "ID", jobID)
			}
		}
	}()

}

func doCancelJob(jobID int) error {
	slog.Info("开始执行每日预约处理程序", "ID", jobID)
	ctx := context.Background()
	err := server.DB.DoQuery(context.Background(), "system", func(q *sqlc.Queries) error {
		//1.获取今日（实际上获取所有过去的预约单来保险）预约
		allZone, _ := model.BlocksInZone("all")
		var beforeScheduledTickets = []sqlc.WtsVTicket{}
		t, err := q.FilterTickets(ctx, sqlc.FilterTicketsParams{
			Blocks: allZone,
			Status: []sqlc.WtsStatus{"scheduled"},
		})
		if err != nil {
			return fmt.Errorf("在获取工单时失败了：%w", err)
		}
		for _, a := range t {
			now := time.Now()

			var date time.Time
			if a.AppointedAt.Valid {
				date = a.AppointedAt.Time
			} else {
				continue
			}

			if date.Before(now) {
				beforeScheduledTickets = append(beforeScheduledTickets, a)
			}
		}
		//2.将所有工单改为“已取消”
		var ticketIDSlice []int32
		for _, a := range beforeScheduledTickets {
			ticketIDSlice = append(ticketIDSlice, a.Tid)
		}
		slog.Info("本次操作共涉及如下工单", "t", ticketIDSlice, "ID", jobID)
		var noErr = true
		var result []int32
		for _, a := range ticketIDSlice {
			t, err := q.CreateTicketTrace(ctx, sqlc.CreateTicketTraceParams{
				Tid: a,
				UpdatedAt: pgtype.Timestamptz{
					Time:  time.Now(),
					Valid: true,
				},
				Op: "-1",
				NewStatus: sqlc.NullWtsStatus{
					WtsStatus: "canceled",
					Valid:     true,
				},
				NewPriority: sqlc.NullWtsPriority{
					WtsPriority: "mainline",
					Valid:       true,
				},
				Remark: "系统检测到预约已过期，似乎是我们爽约了，我们非常抱歉为您造成的不便，您可以再次提交报修预约，我们会努力做得更好。",
			})
			if err != nil {
				noErr = false
				slog.Error("自动预约处理程序增添trace时失败", "error", err)
				continue
			}
			result = append(result, t.Tid)
		}
		slog.Info("本次操作实际操作的工单", "t", result, "ID", jobID)
		//3.如果没有问题就提交事务
		if !noErr {
			return errors.New("增加trace时出现错误，请查看日志。")
		}
		return nil
	})
	return err
}
