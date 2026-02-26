package model

import (
	"errors"

	"zsxyww.com/wts/model/sqlc"
)

var zone = map[string][]sqlc.WtsBlock{
	"FX":    {sqlc.WtsBlock1, sqlc.WtsBlock2, sqlc.WtsBlock3, sqlc.WtsBlock4, sqlc.WtsBlock5, sqlc.WtsBlock6},
	"BM":    {sqlc.WtsBlock7, sqlc.WtsBlock8, sqlc.WtsBlock9, sqlc.WtsBlock10, sqlc.WtsBlock11},
	"DM":    {sqlc.WtsBlock12, sqlc.WtsBlock13, sqlc.WtsBlock14, sqlc.WtsBlock15, sqlc.WtsBlock20, sqlc.WtsBlock21, sqlc.WtsBlock22},
	"QT":    {sqlc.WtsBlock16, sqlc.WtsBlock17, sqlc.WtsBlock18, sqlc.WtsBlock19},
	"XHAB":  {sqlc.WtsBlockXHA, sqlc.WtsBlockXHB},
	"XHCD":  {sqlc.WtsBlockXHC, sqlc.WtsBlockXHD},
	"ZH":    {sqlc.WtsBlockZH},
	"other": {sqlc.WtsBlockOther},
	"all":   {sqlc.WtsBlock1, sqlc.WtsBlock2, sqlc.WtsBlock3, sqlc.WtsBlock4, sqlc.WtsBlock5, sqlc.WtsBlock6, sqlc.WtsBlock7, sqlc.WtsBlock8, sqlc.WtsBlock9, sqlc.WtsBlock10, sqlc.WtsBlock11, sqlc.WtsBlock12, sqlc.WtsBlock13, sqlc.WtsBlock14, sqlc.WtsBlock15, sqlc.WtsBlock20, sqlc.WtsBlock21, sqlc.WtsBlock22, sqlc.WtsBlock16, sqlc.WtsBlock17, sqlc.WtsBlock18, sqlc.WtsBlock19, sqlc.WtsBlockXHA, sqlc.WtsBlockXHB, sqlc.WtsBlockXHC, sqlc.WtsBlockXHD, sqlc.WtsBlockZH, sqlc.WtsBlockOther},
}

func BlocksInZone(zoneName string) ([]sqlc.WtsBlock, error) {

	a, ok := zone[zoneName]
	if !ok {
		return a, errors.New("unknown zone name: " + zoneName)
	}
	return a, nil
}
