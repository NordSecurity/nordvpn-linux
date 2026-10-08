// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'recommended_server_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(RecommendedServer)
final recommendedServerProvider = RecommendedServerProvider._();

final class RecommendedServerProvider
    extends
        $AsyncNotifierProvider<RecommendedServer, RecommendedServerLocation> {
  RecommendedServerProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'recommendedServerProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$recommendedServerHash();

  @$internal
  @override
  RecommendedServer create() => RecommendedServer();
}

String _$recommendedServerHash() => r'6ea92dfb0a20221b1a1ea2af2970275cd4cd0ebb';

abstract class _$RecommendedServer
    extends $AsyncNotifier<RecommendedServerLocation> {
  FutureOr<RecommendedServerLocation> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref =
        this.ref
            as $Ref<
              AsyncValue<RecommendedServerLocation>,
              RecommendedServerLocation
            >;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<
                AsyncValue<RecommendedServerLocation>,
                RecommendedServerLocation
              >,
              AsyncValue<RecommendedServerLocation>,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
