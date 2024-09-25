package services

import "avitoTask/internal/db"

const quorum int = 3

func MakingDecision(bid *db.BidModel, tender *db.TenderModel, bidId, tenderId, username, decision string) error {
	tx, err := bid.BeginTransaction()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = bid.CreateDecision(tx, bidId, username, decision)
	if err != nil {
		return err
	}

	if decision == "Rejected" {
		err = bid.ChangeDecision(tx, decision, bidId)
		if err != nil {
			return err
		}
	} else {
		var decisionCnt int
		err = bid.GetApprovedCount(tx, &decisionCnt, bidId)
		if err != nil {
			return err
		}
		if decisionCnt >= quorum {
			err = bid.ChangeDecision(tx, decision, bidId)
			if err != nil {
				return err
			}
			err = tender.ChangeStatus("Closed", tenderId, tx)
			if err != nil {
				return err
			}
		}
	}
	tx.Commit()

	return nil
}
