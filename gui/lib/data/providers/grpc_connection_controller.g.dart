// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'grpc_connection_controller.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(GrpcConnectionController)
const grpcConnectionControllerProvider = GrpcConnectionControllerProvider._();

final class GrpcConnectionControllerProvider
    extends $AsyncNotifierProvider<GrpcConnectionController, bool> {
  const GrpcConnectionControllerProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'grpcConnectionControllerProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$grpcConnectionControllerHash();

  @$internal
  @override
  GrpcConnectionController create() => GrpcConnectionController();
}

String _$grpcConnectionControllerHash() =>
    r'0ace884bdb389eafb4f056f67bdd76c9dca13ce6';

abstract class _$GrpcConnectionController extends $AsyncNotifier<bool> {
  FutureOr<bool> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final created = build();
    final ref = this.ref as $Ref<AsyncValue<bool>, bool>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<bool>, bool>,
              AsyncValue<bool>,
              Object?,
              Object?
            >;
    element.handleValue(ref, created);
  }
}
