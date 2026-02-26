package logic

import (
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/model"
	"zsxyww.com/wts/model/sqlc"
)

func TicketOverview(c *hutil.WtsCtx, op string) hutil.TicketOverviewResponse {

	ctx := c.Request().Context()

	var result hutil.TicketOverviewResponse

	err := c.DB.DoQuery(ctx, op, func(q *sqlc.Queries) error {
		count, err := q.GetActiveTicketCountByBlock(ctx)
		if err != nil {
			return hutil.NewUnknownErr(err)
		}

		//println("")
		//fmt.Printf("%v ", count)
		//println("")

		resultMap := make(map[sqlc.WtsBlock]int64)

		allZone, _ := model.BlocksInZone("all") //大概不会有问题
		for _, a := range allZone {
			resultMap[a] = 0
		}

		for _, a := range count {
			if a.Block.Valid { //应该不会有没有信息的情况，除非用户是通过数据库直接插进来的，不过还是要严谨一点
				resultMap[a.Block.WtsBlock] = a.Total
			}
		}
		result.CountByBlock = resultMap
		return nil
	})
	result.Err = err
	return result

}
