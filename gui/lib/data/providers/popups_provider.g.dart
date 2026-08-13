// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'popups_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(Popups)
final popupsProvider = PopupsProvider._();

final class PopupsProvider extends $NotifierProvider<Popups, PopupMetadata?> {
  PopupsProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'popupsProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$popupsHash();

  @$internal
  @override
  Popups create() => Popups();

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(PopupMetadata? value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<PopupMetadata?>(value),
    );
  }
}

String _$popupsHash() => r'fd8a7ba721b83574d4e3f73e62c2bcdc9b06cac3';

abstract class _$Popups extends $Notifier<PopupMetadata?> {
  PopupMetadata? build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref = this.ref as $Ref<PopupMetadata?, PopupMetadata?>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<PopupMetadata?, PopupMetadata?>,
              PopupMetadata?,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
