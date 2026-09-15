// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'vpn_status_controller.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning
/// Handles the VPN connection functionality

@ProviderFor(VpnStatusController)
final vpnStatusControllerProvider = VpnStatusControllerProvider._();

/// Handles the VPN connection functionality
final class VpnStatusControllerProvider
    extends $AsyncNotifierProvider<VpnStatusController, VpnStatus> {
  /// Handles the VPN connection functionality
  VpnStatusControllerProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'vpnStatusControllerProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$vpnStatusControllerHash();

  @$internal
  @override
  VpnStatusController create() => VpnStatusController();
}

String _$vpnStatusControllerHash() =>
    r'88f54177b7009d681b13542cea5a8a5e39924569';

/// Handles the VPN connection functionality

abstract class _$VpnStatusController extends $AsyncNotifier<VpnStatus> {
  FutureOr<VpnStatus> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref = this.ref as $Ref<AsyncValue<VpnStatus>, VpnStatus>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<VpnStatus>, VpnStatus>,
              AsyncValue<VpnStatus>,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
