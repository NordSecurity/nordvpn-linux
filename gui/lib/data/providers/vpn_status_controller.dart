import 'package:grpc/grpc.dart' hide ConnectionState;
import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/pause.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/app_state_provider.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/providers/toasts_provider.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/data/repository/vpn_repository.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/state.pbenum.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';
import 'package:nordvpn/pb/daemon/uievent.pbenum.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'vpn_status_controller.g.dart';

typedef _VpnRepoFn = Future<int> Function(VpnRepository);

/// Handles the VPN connection functionality
@riverpod
class VpnStatusController extends _$VpnStatusController
    implements VpnStatusObserver, PauseEventsObserver {
  @override
  FutureOr<VpnStatus> build() async {
    final status = await ref.read(vpnRepositoryProvider).fetchStatus();
    _registerNotifications();

    if (status.state == ConnectionState.PAUSED) {
      // show the toast if the gui starts while already paused
      ref
          .read(toastsProvider.notifier)
          .show(Duration(seconds: status.pauseRemainingDurationSec));
    } else {
      // clear any stale toast left over from a previous paused session
      ref.read(toastsProvider.notifier).closeToast();
    }

    return VpnStatus.fromStatusResponse(status);
  }

  /// Connect to a VPN server.
  ///
  /// For regular connections from the server list, quick connect, etc.
  Future<void> connect(ConnectArguments? args) {
    return _doAndShowPopup(
      (vpn) => vpn.connect(
        args ?? ConnectArguments(),
        itemName: UIEvent_ItemName.CONNECT,
      ),
    );
  }

  /// Connect from the Recent Connections list.
  Future<void> connectFromRecents(ConnectArguments args) {
    return _doAndShowPopup(
      (vpn) => vpn.connect(args, itemName: UIEvent_ItemName.CONNECT_RECENTS),
    );
  }

  Future<void> reconnect(ConnectionParameters args) =>
      _doAndShowPopup((vpn) => vpn.reconnect(args));

  /// Reconnect using [args]; if no suitable server exists for the currently
  /// selected technology (e.g. Double VPN has no NordWhisper servers), fall
  /// back to a quick connect so the newly chosen technology is still applied.
  Future<void> reconnectOrQuickConnect(ConnectionParameters args) async {
    final status = await _run((vpn) => vpn.reconnect(args));
    if (status == DaemonStatusCode.serverUnavailable) {
      logger.i(
        "no suitable server for reconnect params, falling back to quick connect",
      );
      await connect(null); // quick connect for the current (new) technology
      return;
    }
    ref.read(popupsProvider.notifier).show(status);
  }

  Future<void> disconnect() => _doAndShowPopup((vpn) => vpn.disconnect());

  Future<void> pauseConnection(PauseLength pauseValue) =>
      _doAndShowPopup((vpn) => vpn.pauseConnection(pauseValue));

  Future<int> cancelConnect() => _doAndShowPopup((vpn) => vpn.cancelConnect());

  Future<int> _doAndShowPopup(_VpnRepoFn fn) async {
    final status = await _run(fn);
    ref.read(popupsProvider.notifier).show(status);
    return status;
  }

  /// Runs the repo call and resolves it to a [DaemonStatusCode] without showing
  /// any popup. Use [_doAndShowPopup] when the result should also be surfaced.
  Future<int> _run(_VpnRepoFn fn) async {
    final vpnProvider = ref.read(vpnRepositoryProvider);
    int status = DaemonStatusCode.success;

    try {
      status = await fn(vpnProvider);
    } on GrpcError catch (e) {
      // Convert gRPC Error into DaemonStatusCode
      status = DaemonStatusCode.fromGrpcError(e);
    } catch (e) {
      // Unexpected error during the gRPC call
      status = DaemonStatusCode.failure;
      // Log the error
      logger.e("Unexpected error: $e");
    }

    if (status == DaemonStatusCode.failure) {
      status = DaemonStatusCode.failedToConnectToVpn;
    }

    return status;
  }

  void _registerNotifications() {
    final notification = ref.read(appStateProvider);
    notification.addVpnStatusObserver(this);
    notification.addPauseEventsObserver(this);
    ref.onDispose(() {
      notification.removePauseEventsObserver(this);
      notification.removeVpnStatusObserver(this);
    });
  }

  @override
  void onVpnStatusChanged(StatusResponse status) {
    if (!state.hasValue) {
      // if there is no status already fetched it is not possible to construct the status
      logger.d("ignore VPN status changed because app doesn't have a status");
      return;
    }

    if (state.value!.isEqualToStatusResponse(status)) {
      // if the new value is the same ignore, otherwise widgets are recreated
      logger.d(
        "ignore VPN status changed ${status.state} because they are equal",
      );
      return;
    }

    if (status.state == ConnectionState.PAUSED) {
      ref
          .read(toastsProvider.notifier)
          .show(Duration(seconds: status.pauseRemainingDurationSec));
    } else {
      ref.read(toastsProvider.notifier).closeToast();
    }

    final vpnStatus = state.value!.copyWith(
      ip: status.ip.isNotEmpty ? status.ip : null,
      hostname: status.hostname.isNotEmpty ? status.hostname : null,
      country: status.hasCountry() ? Country.fromCode(status.country) : null,
      city: status.city.isNotEmpty ? City(status.city) : null,
      status: status.state,
      isVirtualLocation: status.virtualLocation,
      isObfuscated: status.obfuscated,
      connectionParameters: status.parameters,
      isMeshnetRouting: status.isMeshPeer,
    );

    state = AsyncData(vpnStatus);
  }

  @override
  void onPauseEvent(PauseEventType type) {
    var status = DaemonStatusCode.success;

    if (type == PauseEventType.RECONNECT_FAILED) {
      status = DaemonStatusCode.failedToConnectToVpn;
    }

    ref.read(popupsProvider.notifier).show(status);
  }
}
