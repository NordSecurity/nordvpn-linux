import 'package:grpc/grpc.dart';
import 'package:nordvpn/data/models/city.dart';
import 'package:nordvpn/data/models/connect_arguments.dart';
import 'package:nordvpn/data/models/country.dart';
import 'package:nordvpn/data/models/vpn_status.dart';
import 'package:nordvpn/data/providers/app_state_provider.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/data/repository/vpn_repository.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/pb/daemon/status.pb.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'vpn_status_controller.g.dart';

typedef _VpnRepoFn = Future<int> Function(VpnRepository);

// Handles the VPN connection functionality
@riverpod
class VpnStatusController extends _$VpnStatusController
    implements VpnStatusObserver {
  @override
  FutureOr<VpnStatus> build() async {
    final status = await ref.read(vpnRepositoryProvider).fetchStatus();
    _registerNotifications();

    return VpnStatus.fromStatusResponse(status);
  }

  Future<void> connect(ConnectArguments? args) =>
      _doAndShowPopup((vpn) => vpn.connect(args ?? ConnectArguments()));

  Future<void> reconnect(ConnectionParameters args) =>
      _doAndShowPopup((vpn) => vpn.reconnect(args));

  Future<void> disconnect() => _doAndShowPopup((vpn) => vpn.disconnect());

  Future<int> cancelConnect() => _doAndShowPopup((vpn) => vpn.cancelConnect());

  Future<int> _doAndShowPopup(_VpnRepoFn fn) async {
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

    ref.read(popupsProvider.notifier).show(status);
    return status;
  }

  void _registerNotifications() {
    final notification = ref.read(appStateProvider);
    notification.addVpnStatusObserver(this);
    ref.onDispose(() {
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

    final vpnStatus = state.value!.copyWith(
      ip: status.ip.isNotEmpty ? status.ip : null,
      hostname: status.hostname.isNotEmpty ? status.hostname : null,
      country: status.hasCountry() ? Country.fromCode(status.country) : null,
      city: status.city.isNotEmpty ? City(status.city) : null,
      status: status.state,
      isVirtualLocation: status.virtualLocation,
      connectionParameters: status.parameters,
      isMeshnetRouting: status.isMeshPeer,
    );

    state = AsyncData(vpnStatus);
  }
}
