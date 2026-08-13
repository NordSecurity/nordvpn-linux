// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'token_info_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(TokenInfo)
final tokenInfoProvider = TokenInfoProvider._();

final class TokenInfoProvider
    extends $AsyncNotifierProvider<TokenInfo, TokenInfoResponse?> {
  TokenInfoProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'tokenInfoProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$tokenInfoHash();

  @$internal
  @override
  TokenInfo create() => TokenInfo();
}

String _$tokenInfoHash() => r'c84ab93770116ddb78911d4aea328e9d118de23d';

abstract class _$TokenInfo extends $AsyncNotifier<TokenInfoResponse?> {
  FutureOr<TokenInfoResponse?> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref =
        this.ref as $Ref<AsyncValue<TokenInfoResponse?>, TokenInfoResponse?>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<TokenInfoResponse?>, TokenInfoResponse?>,
              AsyncValue<TokenInfoResponse?>,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
