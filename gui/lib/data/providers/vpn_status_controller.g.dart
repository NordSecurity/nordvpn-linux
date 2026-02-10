// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'vpn_status_controller.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(VpnStatusController)
final vpnStatusControllerProvider = VpnStatusControllerProvider._();

final class VpnStatusControllerProvider
    extends $AsyncNotifierProvider<VpnStatusController, VpnStatus> {
  VpnStatusControllerProvider._()
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
    r'1c8b427f411be47b40473ddcd66f747b9f43cf00';

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
