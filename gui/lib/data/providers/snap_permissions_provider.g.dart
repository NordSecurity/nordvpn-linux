// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'snap_permissions_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(SnapPermissions)
final snapPermissionsProvider = SnapPermissionsProvider._();

final class SnapPermissionsProvider
    extends $AsyncNotifierProvider<SnapPermissions, List<String>> {
  SnapPermissionsProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'snapPermissionsProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$snapPermissionsHash();

  @$internal
  @override
  SnapPermissions create() => SnapPermissions();
}

String _$snapPermissionsHash() => r'c6d2c22713b7f593f7ef3efd39916d468febb370';

abstract class _$SnapPermissions extends $AsyncNotifier<List<String>> {
  FutureOr<List<String>> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref = this.ref as $Ref<AsyncValue<List<String>>, List<String>>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<List<String>>, List<String>>,
              AsyncValue<List<String>>,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
