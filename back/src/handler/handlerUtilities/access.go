package hutil

import (
	"slices"

	"zsxyww.com/wts/model/sqlc"
)

// usage: IsAccessIn("api", "group-leader")("user") -> false
//
// 所有可以访问API的权限都要被显式地列出。比如，如果你设置`IsAccessIn("api","formal-member")("chief")`，那么只有API和正式成员才能访问，就是科长也不行。
//
// 系统的权限就那么多，我觉得这样做是完全合理的，列出所有权限也更加清晰且安全。
func IsAccessIn(targets ...sqlc.WtsAccess) func(subject sqlc.WtsAccess) bool {
	return func(subject sqlc.WtsAccess) bool {
		return slices.Contains(targets, subject)
	}
}

//也可以调用下面的函数

var IsOperator = IsAccessIn(
	sqlc.WtsAccessApi,
	sqlc.WtsAccessChief,
	sqlc.WtsAccessDev,
	sqlc.WtsAccessGroupLeader,
	sqlc.WtsAccessFormalMember,
	sqlc.WtsAccessInformalMember)

var IsAdmin = IsAccessIn(
	sqlc.WtsAccessGroupLeader,
	sqlc.WtsAccessApi,
	sqlc.WtsAccessChief,
	sqlc.WtsAccessDev)

var IsUser = IsAccessIn(
	sqlc.WtsAccessApi,
	sqlc.WtsAccessChief,
	sqlc.WtsAccessDev,
	sqlc.WtsAccessGroupLeader,
	sqlc.WtsAccessFormalMember,
	sqlc.WtsAccessInformalMember,
	sqlc.WtsAccessPreMember,
	sqlc.WtsAccessUser)

var IsPreMember = IsAccessIn(
	sqlc.WtsAccessPreMember)

var IsFormalMember = IsAccessIn(
	sqlc.WtsAccessGroupLeader,
	sqlc.WtsAccessApi,
	sqlc.WtsAccessChief,
	sqlc.WtsAccessDev,
	sqlc.WtsAccessFormalMember)

var IsUnregistered = IsAccessIn(
	sqlc.WtsAccessUnregistered)
