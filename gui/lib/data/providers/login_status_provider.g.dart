// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'login_status_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(LoginStatus)
final loginStatusProvider = LoginStatusProvider._();

final class LoginStatusProvider
    extends $AsyncNotifierProvider<LoginStatus, bool> {
  LoginStatusProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'loginStatusProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$loginStatusHash();

  @$internal
  @override
  LoginStatus create() => LoginStatus();
}

String _$loginStatusHash() => r'9d57d3a425db3ba7f3d7c5c400f776219a34d972';

abstract class _$LoginStatus extends $AsyncNotifier<bool> {
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
