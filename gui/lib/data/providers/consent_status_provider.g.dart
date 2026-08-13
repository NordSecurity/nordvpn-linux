// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'consent_status_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(ConsentStatus)
const consentStatusProvider = ConsentStatusProvider._();

final class ConsentStatusProvider
    extends $AsyncNotifierProvider<ConsentStatus, ConsentLevel> {
  const ConsentStatusProvider._()
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

String _$consentStatusHash() => r'bed845c5e40724cf447addf3896ba600fc749f0d';

abstract class _$ConsentStatus extends $AsyncNotifier<ConsentLevel> {
  FutureOr<ConsentLevel> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final created = build();
    final ref = this.ref as $Ref<AsyncValue<ConsentLevel>, ConsentLevel>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<ConsentLevel>, ConsentLevel>,
              AsyncValue<ConsentLevel>,
              Object?,
              Object?
            >;
    element.handleValue(ref, created);
  }
}
