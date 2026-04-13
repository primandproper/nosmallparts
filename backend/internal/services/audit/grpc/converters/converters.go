package converters

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/converters"
	auditsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/audit"
	"github.com/verygoodsoftwarenotvirus/platform/v4/pointer"
)

func ConvertAuditLogEntryToGRPCAuditLogEntry(entry *audit.AuditLogEntry) *auditsvc.AuditLogEntry {
	changes := map[string]*auditsvc.ChangeLog{}
	for k, v := range entry.Changes {
		changes[k] = &auditsvc.ChangeLog{
			OldValue: v.OldValue,
			NewValue: v.NewValue,
		}
	}

	return &auditsvc.AuditLogEntry{
		CreatedAt:        grpcconverters.ConvertTimeToPBTimestamp(entry.CreatedAt),
		Changes:          changes,
		BelongsToAccount: pointer.Dereference(entry.BelongsToAccount),
		Id:               entry.ID,
		ResourceType:     entry.ResourceType,
		RelevantId:       entry.RelevantID,
		EventType:        entry.EventType,
		BelongsToUser:    entry.BelongsToUser,
	}
}
