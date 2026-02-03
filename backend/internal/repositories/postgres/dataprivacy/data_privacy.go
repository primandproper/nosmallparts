package dataprivacy

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/dataprivacy"
)

func (r *repository) FetchUserDataCollection(ctx context.Context, userID string) (*dataprivacy.UserDataCollectionResponse, error) {
	_, span := r.tracer.StartSpan(ctx)
	defer span.End()

	/*
		NOTE: none of this is how it's meant to be, I just haven't implemented it properly yet. The code below is to make the linter stop whining.
	*/

	accountID := "unknown"
	accounts, err := r.identityRepo.GetAccounts(ctx, userID, nil)
	if err != nil {
		return nil, err
	}
	if len(accounts.Data) > 0 {
		accountID = accounts.Data[0].ID
	}

	x := &dataprivacy.UserDataCollectionResponse{
		ReportID: accountID,
	}

	r.logger.Info("TODO: FetchUserDataCollection")

	return x, nil
}

func (r *repository) DeleteUser(ctx context.Context, userID string) error {
	// TODO: implement
	return nil
}
