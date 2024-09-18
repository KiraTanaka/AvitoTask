package db

import (
	_ "embed"
)

//go:embed queries/checkBidExists.sql
var checkBidExistsQuery string

//go:embed queries/checkUserCanManageBid.sql
var checkUserCanManageBidQuery string

//go:embed queries/checkUserViewBid.sql
var checkUserViewBidQuery string

//go:embed queries/checkUserCanApproveBid.sql
var checkUserCanApproveBidQuery string

func CheckBidExists(bidId string) error {
	var bidExists bool
	return db.Get(&bidExists, checkBidExistsQuery, bidId)
}

func CheckUserCanManageBid(username, autorType, authorId string) error {
	var canManage bool
	return db.Get(&canManage, checkUserCanManageBidQuery, username, authorId, autorType)
}
func CheckUserViewBid(username, bidId string) error {
	var canView bool
	return db.Get(&canView, checkUserViewBidQuery, bidId, username)
}
func CheckUserCanApproveBid(username, tenderId string) error {
	var canManage bool
	return db.Get(&canManage, checkUserCanApproveBidQuery, tenderId, username)
}
