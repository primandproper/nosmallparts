package helpers

import (
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"
	g "maragu.dev/gomponents"
)

func RenderTimestamp(value any) g.Node {
	if value == nil {
		return g.Text("-")
	}

	switch v := value.(type) {
	case *timestamppb.Timestamp:
		if v == nil {
			return g.Text("-")
		}
		return g.Text(v.AsTime().Format("2006-01-02 15:04:05"))
	case timestamppb.Timestamp:
		return g.Text(v.AsTime().Format("2006-01-02 15:04:05"))
	default:
		return g.Text(fmt.Sprintf("%v", v))
	}
}
