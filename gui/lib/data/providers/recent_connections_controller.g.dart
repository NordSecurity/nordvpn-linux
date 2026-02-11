// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'recent_connections_controller.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(RecentConnectionsController)
final recentConnectionsControllerProvider =
    RecentConnectionsControllerProvider._();

final class RecentConnectionsControllerProvider
    extends
        $AsyncNotifierProvider<
          RecentConnectionsController,
          List<RecentConnection>
        > {
  RecentConnectionsControllerProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'recentConnectionsControllerProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$recentConnectionsControllerHash();

  @$internal
  @override
  RecentConnectionsController create() => RecentConnectionsController();
}

String _$recentConnectionsControllerHash() =>
    r'f91af6069d549ae437cf284d1c0dbbfe6c5b9556';

abstract class _$RecentConnectionsController
    extends $AsyncNotifier<List<RecentConnection>> {
  FutureOr<List<RecentConnection>> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref =
        this.ref
            as $Ref<AsyncValue<List<RecentConnection>>, List<RecentConnection>>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<
                AsyncValue<List<RecentConnection>>,
                List<RecentConnection>
              >,
              AsyncValue<List<RecentConnection>>,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
