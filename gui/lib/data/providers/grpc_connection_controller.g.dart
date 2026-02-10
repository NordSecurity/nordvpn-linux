// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'grpc_connection_controller.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(GrpcConnectionController)
final grpcConnectionControllerProvider = GrpcConnectionControllerProvider._();

final class GrpcConnectionControllerProvider
    extends $AsyncNotifierProvider<GrpcConnectionController, bool> {
  GrpcConnectionControllerProvider._()
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
    r'031d234165569af2bc933993b1ae1c54f3d74a8e';

abstract class _$GrpcConnectionController extends $AsyncNotifier<bool> {
  FutureOr<bool> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref = this.ref as $Ref<AsyncValue<bool>, bool>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<bool>, bool>,
              AsyncValue<bool>,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
