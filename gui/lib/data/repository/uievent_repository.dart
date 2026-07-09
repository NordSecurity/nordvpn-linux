import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/grpc/grpc_service.dart';
import 'package:nordvpn/grpc/uievent_reporter.dart';
import 'package:nordvpn/pb/daemon/service.pbgrpc.dart';
import 'package:nordvpn/pb/daemon/uievent.pbenum.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'uievent_repository.g.dart';

/// Reports analytics-only UI events to the daemon.
/// For events that are not paired with any domain operation
/// (e.g., navigation clicks, help link clicks).
class UiEventRepository {
  final DaemonClient _client;

  UiEventRepository([DaemonClient? client])
    : _client = client ?? createDaemonClient();

  void reportChangeSettings() => reportUIEvent(
    _client,
    formReference: UIEvent_FormReference.CONNECTION_INFO,
    itemName: UIEvent_ItemName.CHANGE_SETTINGS,
  );

  void reportGetHelp() => reportUIEvent(
    _client,
    formReference: UIEvent_FormReference.CONNECTION_INFO,
    itemName: UIEvent_ItemName.GET_HELP,
  );
}

@Riverpod(keepAlive: true)
UiEventRepository uiEventRepository(Ref ref) {
  return UiEventRepository();
}
