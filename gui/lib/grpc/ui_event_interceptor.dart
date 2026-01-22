import 'package:grpc/grpc.dart';
import 'package:nordvpn/pb/daemon/uievent.pbenum.dart';

/// Metadata keys for UI event context passed via gRPC metadata.
/// These must match the Go server-side constants in uievent/context.go.
const String metadataKeyFormReference = 'ui-form-reference';
const String metadataKeyItemName = 'ui-item-name';
const String metadataKeyItemType = 'ui-item-type';
const String metadataKeyItemValue = 'ui-item-value';

/// Creates CallOptions with UI event metadata for gRPC calls.
///
/// This is the way to attach UI event tracking metadata to gRPC calls.
/// Every tracked gRPC call (connect, disconnect, login, logout...) should use this
/// to create CallOptions with the appropriate metadata.
CallOptions createUiEventCallOptions({
  required UIEvent_FormReference formReference,
  required UIEvent_ItemName itemName,
  UIEvent_ItemType itemType = UIEvent_ItemType.CLICK,
  UIEvent_ItemValue? itemValue,
}) {
  final metadata = <String, String>{
    metadataKeyFormReference: formReference.value.toString(),
    metadataKeyItemName: itemName.value.toString(),
    metadataKeyItemType: itemType.value.toString(),
  };

  if (itemValue != null &&
      itemValue != UIEvent_ItemValue.ITEM_VALUE_UNSPECIFIED) {
    metadata[metadataKeyItemValue] = itemValue.value.toString();
  }

  return CallOptions(metadata: metadata);
}
