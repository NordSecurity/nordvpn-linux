// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'consent_status_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(ConsentStatus)
final consentStatusProvider = ConsentStatusProvider._();

final class ConsentStatusProvider
    extends $AsyncNotifierProvider<ConsentStatus, ConsentLevel> {
  ConsentStatusProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'consentStatusProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$consentStatusHash();

  @$internal
  @override
  ConsentStatus create() => ConsentStatus();
}

String _$consentStatusHash() => r'38fbe4b40cf9b9b876b3bf0b6f7aebe82fb36e5d';

abstract class _$ConsentStatus extends $AsyncNotifier<ConsentLevel> {
  FutureOr<ConsentLevel> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref = this.ref as $Ref<AsyncValue<ConsentLevel>, ConsentLevel>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<ConsentLevel>, ConsentLevel>,
              AsyncValue<ConsentLevel>,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
