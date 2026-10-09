import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nordvpn/data/models/vpn_protocol.dart';
import 'package:nordvpn/data/providers/pending_settings_provider.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/providers/vpn_settings_controller.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';

import '../../utils/provider_fakes.dart';

Future<(ProviderContainer, FakeVpnSettingsRepository)> _setup(
  int setVpnProtocolStatus,
) async {
  final repository = FakeVpnSettingsRepository(
    setVpnProtocolStatus: setVpnProtocolStatus,
  );
  final container = createContainer(vpnSettingsRepository: repository);
  await container.read(vpnSettingsControllerProvider.future);
  return (container, repository);
}

Future<bool> _applyPending(ProviderContainer container, VpnProtocol protocol) {
  container.read(pendingVPNProtocolProvider.notifier).set(protocol);
  return container
      .read(vpnSettingsControllerProvider.notifier)
      .applyPendingVPNProtocol();
}

void main() {
  for (final status in [
    DaemonStatusCode.success,
    DaemonStatusCode.successReconnectRequired,
    DaemonStatusCode.nothingToDo,
  ]) {
    test(
      'applyPendingVPNProtocol succeeds without a popup on $status',
      () async {
        final (container, repository) = await _setup(status);

        final applied = await _applyPending(container, VpnProtocol.openVpnTcp);

        expect(applied, isTrue);
        expect(repository.requested, VpnProtocol.openVpnTcp);
        expect(container.read(pendingVPNProtocolProvider), isNull);
        expect(container.read(popupsProvider), isNull);
      },
    );
  }

  test('applyPendingVPNProtocol fails and shows the daemon error', () async {
    final (container, _) = await _setup(
      DaemonStatusCode.dedicatedServersNoNordlynx,
    );

    final applied = await _applyPending(container, VpnProtocol.openVpnUdp);

    expect(applied, isFalse);
    expect(
      container.read(popupsProvider)?.id,
      DaemonStatusCode.dedicatedServersNoNordlynx,
    );
  });

  test('applyPendingVPNProtocol fails when nothing is pending', () async {
    final (container, repository) = await _setup(DaemonStatusCode.success);

    final applied = await container
        .read(vpnSettingsControllerProvider.notifier)
        .applyPendingVPNProtocol();

    expect(applied, isFalse);
    expect(repository.requested, isNull);
  });
}
