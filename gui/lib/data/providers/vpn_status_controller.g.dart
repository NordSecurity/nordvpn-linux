// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'vpn_status_controller.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning
/// Handles the VPN connection functionality

@ProviderFor(VpnStatusController)
const vpnStatusControllerProvider = VpnStatusControllerProvider._();

/// Handles the VPN connection functionality
final class VpnStatusControllerProvider
    extends $AsyncNotifierProvider<VpnStatusController, VpnStatus> {
  /// Handles the VPN connection functionality
  const VpnStatusControllerProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'vpnStatusControllerProvider',
        isAutoDispose: true,
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
    r'eaf4075751e338ebb5eb3d2c388fefb199581694';

/// Handles the VPN connection functionality

abstract class _$VpnStatusController extends $AsyncNotifier<VpnStatus> {
  FutureOr<VpnStatus> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final created = build();
    final ref = this.ref as $Ref<AsyncValue<VpnStatus>, VpnStatus>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<VpnStatus>, VpnStatus>,
              AsyncValue<VpnStatus>,
              Object?,
              Object?
            >;
    element.handleValue(ref, created);
  }
}
