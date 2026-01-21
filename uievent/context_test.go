package uievent

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestToMetadata_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		ctx  *UIEventContext
	}{
		{
			name: "all fields set",
			ctx: &UIEventContext{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_COUNTRY,
			},
		},
		{
			name: "without item value",
			ctx: &UIEventContext{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_LOGIN,
				ItemType:      pb.UIEvent_CLICK,
				ItemValue:     pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := ToMetadata(tt.ctx)
			result := FromMetadata(md)

			require.NotNil(t, result)
			assert.Equal(t, tt.ctx.FormReference, result.FormReference)
			assert.Equal(t, tt.ctx.ItemName, result.ItemName)
			assert.Equal(t, tt.ctx.ItemType, result.ItemType)
			assert.Equal(t, tt.ctx.ItemValue, result.ItemValue)
		})
	}
}

func TestToMetadata_NilContext(t *testing.T) {
	assert.Empty(t, ToMetadata(nil))
}

func TestToMetadata_ItemValueOptional(t *testing.T) {
	ctx := &UIEventContext{
		FormReference: pb.UIEvent_CLI,
		ItemName:      pb.UIEvent_CONNECT,
		ItemType:      pb.UIEvent_CLICK,
		ItemValue:     pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
	}

	md := ToMetadata(ctx)

	assert.NotEmpty(t, md.Get(MetadataKeyFormReference))
	assert.NotEmpty(t, md.Get(MetadataKeyItemName))
	assert.NotEmpty(t, md.Get(MetadataKeyItemType))
	assert.Empty(t, md.Get(MetadataKeyItemValue))
}

func TestFromMetadata_NilOrEmptyMetadata(t *testing.T) {
	assert.Nil(t, FromMetadata(nil))
	assert.Nil(t, FromMetadata(metadata.MD{}))
}

func TestFromMetadata_AllUnspecified(t *testing.T) {
	md := metadata.MD{
		MetadataKeyFormReference: []string{"0"},
		MetadataKeyItemName:      []string{"0"},
		MetadataKeyItemType:      []string{"0"},
	}
	assert.Nil(t, FromMetadata(md))
}

func TestFromMetadata_InvalidIntegerValues(t *testing.T) {
	md := metadata.MD{
		MetadataKeyFormReference: []string{"invalid"},
		MetadataKeyItemName:      []string{"1"},
		MetadataKeyItemType:      []string{"1"},
	}

	result := FromMetadata(md)
	// FormReference will be 0 (unspecified) due to parse error
	// But ItemName and ItemType are valid, so context is returned
	require.NotNil(t, result)
	assert.Equal(t, pb.UIEvent_FORM_REFERENCE_UNSPECIFIED, result.FormReference)
	assert.Equal(t, pb.UIEvent_CONNECT, result.ItemName)
	assert.Equal(t, pb.UIEvent_CLICK, result.ItemType)
}

func TestFromMetadata_PartiallySet(t *testing.T) {
	md := metadata.MD{
		MetadataKeyFormReference: []string{"1"}, // CLI
		MetadataKeyItemName:      []string{"0"}, // UNSPECIFIED
		MetadataKeyItemType:      []string{"0"}, // UNSPECIFIED
	}

	result := FromMetadata(md)
	// At least one field is set, so context is returned
	require.NotNil(t, result)
	assert.Equal(t, pb.UIEvent_CLI, result.FormReference)
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *UIEventContext
		expected bool
	}{
		{
			name:     "nil context",
			ctx:      nil,
			expected: false,
		},
		{
			name: "all required fields set",
			ctx: &UIEventContext{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: true,
		},
		{
			name: "form reference unspecified",
			ctx: &UIEventContext{
				FormReference: pb.UIEvent_FORM_REFERENCE_UNSPECIFIED,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: false,
		},
		{
			name: "item name unspecified",
			ctx: &UIEventContext{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_ITEM_NAME_UNSPECIFIED,
				ItemType:      pb.UIEvent_CLICK,
			},
			expected: false,
		},
		{
			name: "item type unspecified",
			ctx: &UIEventContext{
				FormReference: pb.UIEvent_CLI,
				ItemName:      pb.UIEvent_CONNECT,
				ItemType:      pb.UIEvent_ITEM_TYPE_UNSPECIFIED,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsValid(tt.ctx))
		})
	}
}

func TestAttachToOutgoingContext(t *testing.T) {
	uiCtx := &UIEventContext{
		FormReference: pb.UIEvent_CLI,
		ItemName:      pb.UIEvent_CONNECT,
		ItemType:      pb.UIEvent_CLICK,
		ItemValue:     pb.UIEvent_COUNTRY,
	}

	ctx := AttachToOutgoingContext(context.Background(), uiCtx)

	md, ok := metadata.FromOutgoingContext(ctx)
	require.True(t, ok)
	assert.Equal(t, "1", md.Get(MetadataKeyFormReference)[0]) // CLI = 1
	assert.Equal(t, "1", md.Get(MetadataKeyItemName)[0])      // CONNECT = 1
	assert.Equal(t, "1", md.Get(MetadataKeyItemType)[0])      // CLICK = 1
	assert.Equal(t, "1", md.Get(MetadataKeyItemValue)[0])     // COUNTRY = 1
}

func TestAttachToOutgoingContext_NilUIContext(t *testing.T) {
	originalCtx := context.Background()
	assert.Equal(t, originalCtx, AttachToOutgoingContext(originalCtx, nil))
}

func TestFromIncomingContext_NoMetadata(t *testing.T) {
	assert.Nil(t, FromIncomingContext(context.Background()))
}

func TestFromIncomingContext_WithMetadata(t *testing.T) {
	md := metadata.MD{
		MetadataKeyFormReference: []string{"2"}, // TRAY
		MetadataKeyItemName:      []string{"2"}, // CONNECT_RECENTS
		MetadataKeyItemType:      []string{"1"}, // CLICK
		MetadataKeyItemValue:     []string{"2"}, // CITY
	}
	ctx := metadata.NewIncomingContext(context.Background(), md)

	result := FromIncomingContext(ctx)

	require.NotNil(t, result)
	assert.Equal(t, pb.UIEvent_TRAY, result.FormReference)
	assert.Equal(t, pb.UIEvent_CONNECT_RECENTS, result.ItemName)
	assert.Equal(t, pb.UIEvent_CLICK, result.ItemType)
	assert.Equal(t, pb.UIEvent_CITY, result.ItemValue)
}
