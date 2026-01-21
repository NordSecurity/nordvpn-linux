package uievent

import (
	"context"
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

// intToMetadataValue converts an integer to a metadata value slice.
func intToMetadataValue(val int) []string {
	return []string{strconv.Itoa(val)}
}

// ToMetadata converts a UIEventContext to gRPC metadata.
func ToMetadata(ctx *UIEventContext) metadata.MD {
	if ctx == nil {
		return metadata.MD{}
	}

	md := metadata.MD{
		MetadataKeyFormReference: intToMetadataValue(int(ctx.FormReference)),
		MetadataKeyItemName:      intToMetadataValue(int(ctx.ItemName)),
		MetadataKeyItemType:      intToMetadataValue(int(ctx.ItemType)),
	}

	// ItemValue is optional - only include if not unspecified
	if ctx.ItemValue != pb.UIEvent_ITEM_VALUE_UNSPECIFIED {
		md[MetadataKeyItemValue] = intToMetadataValue(int(ctx.ItemValue))
	}

	return md
}

// getMetadataInt extracts an integer value from metadata.
// Returns 0 if the key is missing or the value is not a valid integer.
func getMetadataInt(md metadata.MD, key string) int {
	values := md.Get(key)
	if len(values) == 0 {
		return 0
	}
	val, err := strconv.Atoi(values[0])
	if err != nil {
		return 0
	}
	return val
}

// FromMetadata extracts UI event context from gRPC metadata.
// Returns nil if required metadata keys are missing.
// Invalid integer values are treated as unspecified (0).
func FromMetadata(md metadata.MD) *UIEventContext {
	if md == nil {
		return nil
	}

	formReference := getMetadataInt(md, MetadataKeyFormReference)
	itemName := getMetadataInt(md, MetadataKeyItemName)
	itemType := getMetadataInt(md, MetadataKeyItemType)
	itemValue := getMetadataInt(md, MetadataKeyItemValue)

	// If all required fields are unspecified, treat as no context
	if formReference == int(pb.UIEvent_FORM_REFERENCE_UNSPECIFIED) &&
		itemName == int(pb.UIEvent_ITEM_NAME_UNSPECIFIED) &&
		itemType == int(pb.UIEvent_ITEM_TYPE_UNSPECIFIED) {
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
