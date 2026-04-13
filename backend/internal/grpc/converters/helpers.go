package grpcconverters

import (
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v4/pointer"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertTimeToPBTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func ConvertTimePointerToPBTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func ConvertPBTimestampToTime(t *timestamppb.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.AsTime()
}

func ConvertPBTimestampToTimePointer(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}

	return pointer.To(t.AsTime())
}

func ConvertUint16PointerToUint32Pointer(i *uint16) *uint32 {
	if i == nil {
		return nil
	}
	return pointer.To(uint32(*i))
}

func ConvertUint32PointerToUint16Pointer(i *uint32) *uint16 {
	if i == nil {
		return nil
	}
	return pointer.To(uint16(*i))
}

func ConvertUint8PointerToUint32Pointer(i *uint8) *uint32 {
	if i == nil {
		return nil
	}
	return pointer.To(uint32(*i))
}

func ConvertUint32PointerToUint8Pointer(i *uint32) *uint8 {
	if i == nil {
		return nil
	}
	return pointer.To(uint8(*i))
}
