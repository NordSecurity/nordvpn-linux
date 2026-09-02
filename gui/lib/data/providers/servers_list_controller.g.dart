// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'servers_list_controller.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(ServersListController)
final serversListControllerProvider = ServersListControllerProvider._();

final class ServersListControllerProvider
    extends $AsyncNotifierProvider<ServersListController, ServersList> {
  ServersListControllerProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'serversListControllerProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$serversListControllerHash();

  @$internal
  @override
  ServersListController create() => ServersListController();
}

String _$serversListControllerHash() =>
    r'0c1fb3c96e273e186581b3f2a168fe3b02ab0ad7';

abstract class _$ServersListController extends $AsyncNotifier<ServersList> {
  FutureOr<ServersList> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref = this.ref as $Ref<AsyncValue<ServersList>, ServersList>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<ServersList>, ServersList>,
              AsyncValue<ServersList>,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
