package uievent

import (
	"context"
	"math"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"google.golang.org/grpc/metadata"
)

// Metadata keys for UI event context passed via gRPC metadata
const (
	MetadataKeyFormReference = "ui-form-reference"
	MetadataKeyItemName      = "ui-item-name"
	MetadataKeyItemType      = "ui-item-type"
	MetadataKeyItemValue     = "ui-item-value"
)

// UIEventContext holds the UI event context data extracted from gRPC metadata
type UIEventContext struct {
	FormReference pb.UIEvent_FormReference
	ItemName      pb.UIEvent_ItemName
	ItemType      pb.UIEvent_ItemType
	ItemValue     pb.UIEvent_ItemValue
}

// NewClickContext creates a UIEventContext for a click action.
func NewClickContext(formRef pb.UIEvent_FormReference, itemName pb.UIEvent_ItemName) *UIEventContext {
	return &UIEventContext{
		FormReference: formRef,
		ItemName:      itemName,
		ItemType:      pb.UIEvent_CLICK,
		ItemValue:     pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
	}
}

// int32ToMetadataValue converts an int32 to a metadata value slice.
func int32ToMetadataValue(val int32) []string {
	return []string{strconv.FormatInt(int64(val), 10)}
}

// ToMetadata converts a UIEventContext to gRPC metadata.
func ToMetadata(ctx *UIEventContext) metadata.MD {
	if ctx == nil {
		return metadata.MD{}
	}

	md := metadata.MD{
		MetadataKeyFormReference: int32ToMetadataValue(int32(ctx.FormReference)),
		MetadataKeyItemName:      int32ToMetadataValue(int32(ctx.ItemName)),
		MetadataKeyItemType:      int32ToMetadataValue(int32(ctx.ItemType)),
	}

	// ItemValue is optional - only include if not unspecified
	if ctx.ItemValue != pb.UIEvent_ITEM_VALUE_UNSPECIFIED {
		md[MetadataKeyItemValue] = int32ToMetadataValue(int32(ctx.ItemValue))
	}

	return md
}

// getMetadataInt32 extracts an int32 value from metadata.
// Returns 0 if the key is missing, the value is not a valid integer,
// or the value is outside int32 range.
func getMetadataInt32(md metadata.MD, key string) int32 {
	values := md.Get(key)
	if len(values) == 0 {
		return 0
	}
	val, err := strconv.ParseInt(values[0], 10, 32)
	if err != nil {
		return 0
	}
	if val < math.MinInt32 || val > math.MaxInt32 {
		return 0
	}
	return int32(val)
}

// FromMetadata extracts UI event context from gRPC metadata.
// Returns nil if required metadata keys are missing.
// Invalid integer values are treated as unspecified (0).
func FromMetadata(md metadata.MD) *UIEventContext {
	if md == nil {
		return nil
	}

	formReference := getMetadataInt32(md, MetadataKeyFormReference)
	itemName := getMetadataInt32(md, MetadataKeyItemName)
	itemType := getMetadataInt32(md, MetadataKeyItemType)
	itemValue := getMetadataInt32(md, MetadataKeyItemValue)

	// If all required fields are unspecified, treat as no context
	if formReference == int32(pb.UIEvent_FORM_REFERENCE_UNSPECIFIED) &&
		itemName == int32(pb.UIEvent_ITEM_NAME_UNSPECIFIED) &&
		itemType == int32(pb.UIEvent_ITEM_TYPE_UNSPECIFIED) {
		return nil
	}

	return &UIEventContext{
		FormReference: pb.UIEvent_FormReference(formReference),
		ItemName:      pb.UIEvent_ItemName(itemName),
		ItemType:      pb.UIEvent_ItemType(itemType),
		ItemValue:     pb.UIEvent_ItemValue(itemValue),
	}
}

// IsValid checks if the required fields are set (not UNSPECIFIED).
// FormReference, ItemName, and ItemType are required. ItemValue is optional.
func IsValid(ctx *UIEventContext) bool {
	if ctx == nil {
		return false
	}
	return ctx.FormReference != pb.UIEvent_FORM_REFERENCE_UNSPECIFIED &&
		ctx.ItemName != pb.UIEvent_ITEM_NAME_UNSPECIFIED &&
		ctx.ItemType != pb.UIEvent_ITEM_TYPE_UNSPECIFIED
}

// AttachToOutgoingContext attaches UI event metadata to an outgoing gRPC context.
func AttachToOutgoingContext(ctx context.Context, uiCtx *UIEventContext) context.Context {
	if uiCtx == nil {
		return ctx
	}
	md := ToMetadata(uiCtx)
	return metadata.NewOutgoingContext(ctx, md)
}

// FromIncomingContext extracts UI event context from an incoming gRPC context.
func FromIncomingContext(ctx context.Context) *UIEventContext {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	return FromMetadata(md)
}
