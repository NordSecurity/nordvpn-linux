// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'vpn_settings_controller.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(VpnSettingsController)
final vpnSettingsControllerProvider = VpnSettingsControllerProvider._();

final class VpnSettingsControllerProvider
    extends $AsyncNotifierProvider<VpnSettingsController, ApplicationSettings> {
  VpnSettingsControllerProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'vpnSettingsControllerProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$vpnSettingsControllerHash();

  @$internal
  @override
  VpnSettingsController create() => VpnSettingsController();
}

String _$vpnSettingsControllerHash() =>
    r'813acb002fa2276987d7eab68ba189b0c450d9d7';

abstract class _$VpnSettingsController
    extends $AsyncNotifier<ApplicationSettings> {
  FutureOr<ApplicationSettings> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref =
        this.ref as $Ref<AsyncValue<ApplicationSettings>, ApplicationSettings>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<ApplicationSettings>, ApplicationSettings>,
              AsyncValue<ApplicationSettings>,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
