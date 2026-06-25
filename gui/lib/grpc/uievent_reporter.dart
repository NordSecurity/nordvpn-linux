import 'package:nordvpn/pb/daemon/service.pbgrpc.dart';
import 'package:nordvpn/pb/daemon/uievent.pb.dart';

/// Builds a [UIEvent] protobuf message from the given parameters.
UIEvent buildUIEvent({
  required UIEvent_FormReference formReference,
  required UIEvent_ItemName itemName,
  UIEvent_ItemType itemType = UIEvent_ItemType.CLICK,
  UIEvent_ItemValue? itemValue,
}) {
  final event = UIEvent()
    ..formReference = formReference
    ..itemName = itemName
    ..itemType = itemType;
  if (itemValue != null &&
      itemValue != UIEvent_ItemValue.ITEM_VALUE_UNSPECIFIED) {
    event.itemValue = itemValue;
  }
  return event;
}

/// Sends a UI analytics event to the daemon via the dedicated
/// ReportUIEvent RPC. The call is fire-and-forget.
void reportUIEvent(
  DaemonClient client, {
  required UIEvent_FormReference formReference,
  required UIEvent_ItemName itemName,
  UIEvent_ItemType itemType = UIEvent_ItemType.CLICK,
  UIEvent_ItemValue? itemValue,
}) {
  final event = buildUIEvent(
    formReference: formReference,
    itemName: itemName,
    itemType: itemType,
    itemValue: itemValue,
  );
  client.reportUIEvent(event);
}
