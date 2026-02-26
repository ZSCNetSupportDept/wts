package hutil

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"zsxyww.com/wts/model"
	"zsxyww.com/wts/model/sqlc"
)

func RegisterValidator(v *validator.Validate) {

	// 验证楼号（是不是sqlc.WtsBlock 系列枚举中的一个）
	v.RegisterValidation("isWtsBlock", func(fl validator.FieldLevel) bool {
		block := fl.Field().String()
		switch sqlc.WtsBlock(block) {
		case sqlc.WtsBlock1, sqlc.WtsBlock2, sqlc.WtsBlock3, sqlc.WtsBlock4, sqlc.WtsBlock5,
			sqlc.WtsBlock6, sqlc.WtsBlock7, sqlc.WtsBlock8, sqlc.WtsBlock9, sqlc.WtsBlock10,
			sqlc.WtsBlock11, sqlc.WtsBlock12, sqlc.WtsBlock13, sqlc.WtsBlock14, sqlc.WtsBlock15,
			sqlc.WtsBlock16, sqlc.WtsBlock17, sqlc.WtsBlock18, sqlc.WtsBlock19, sqlc.WtsBlock20,
			sqlc.WtsBlock21, sqlc.WtsBlock22, sqlc.WtsBlockXHA, sqlc.WtsBlockXHB, sqlc.WtsBlockXHC,
			sqlc.WtsBlockXHD, sqlc.WtsBlockZH, sqlc.WtsBlockOther:
			return true
		default:
			return false
		}
	})

	// 验证是不是中国大陆11位的手机号
	v.RegisterValidation("isValidPhone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		// 匹配以1开头的11位数字，第二位是3-9
		re := regexp.MustCompile(`^1[3-9]\d{9}$`)
		return re.MatchString(phone)
	})

	// 验证是不是sqlc.WtsISP系列枚举中的一个
	v.RegisterValidation("isValidISP", func(fl validator.FieldLevel) bool {
		isp := fl.Field().String()
		switch sqlc.WtsIsp(isp) {
		case sqlc.WtsIspTelecom, sqlc.WtsIspUnicom, sqlc.WtsIspMobile, sqlc.WtsIspBroadnet, sqlc.WtsIspOthers:
			return true
		default:
			return false
		}
	})

	// 验证是不是sqlc.WtsPriority系列枚举中的一个
	v.RegisterValidation("isValidPriority", func(fl validator.FieldLevel) bool {
		priority := fl.Field().String()
		switch sqlc.WtsPriority(priority) {
		case sqlc.WtsPriorityHighest, sqlc.WtsPriorityMainline, sqlc.WtsPriorityAssigned, sqlc.WtsPriorityNormal, sqlc.WtsPriorityInPassing, sqlc.WtsPriorityLeast:
			return true
		default:
			return false
		}
	})

	// 验证是不是sqlc.WtsCategory系列枚举中的一个
	v.RegisterValidation("isValidCategory", func(fl validator.FieldLevel) bool {
		category := fl.Field().String()
		switch sqlc.WtsCategory(category) {
		case sqlc.WtsCategoryFirstInstall, sqlc.WtsCategoryClientOrAccount, sqlc.WtsCategoryIpOrDevice, sqlc.WtsCategoryLowSpeed, sqlc.WtsCategoryOthers:
			return true
		default:
			return false
		}
	})

	// 验证是不是sqlc.WtsStatus系列枚举中的一个
	v.RegisterValidation("isValidStatus", func(fl validator.FieldLevel) bool {
		status := fl.Field().String()
		switch sqlc.WtsStatus(status) {
		case sqlc.WtsStatusFresh, sqlc.WtsStatusDelay, sqlc.WtsStatusScheduled, sqlc.WtsStatusCanceled, sqlc.WtsStatusEscalated, sqlc.WtsStatusSolved:
			return true
		default:
			return false
		}
	})

	// 验证是不是有效的片区
	v.RegisterValidation("isValidZone", func(fl validator.FieldLevel) bool {
		zone := fl.Field().String()
		_, err := model.BlocksInZone(zone)
		return err == nil

	})

}
