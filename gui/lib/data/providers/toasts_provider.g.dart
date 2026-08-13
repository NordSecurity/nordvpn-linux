// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'toasts_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(Toasts)
final toastsProvider = ToastsProvider._();

final class ToastsProvider extends $NotifierProvider<Toasts, Duration?> {
  ToastsProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'toastsProvider',
        isAutoDispose: false,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$toastsHash();

  @$internal
  @override
  Toasts create() => Toasts();

  /// {@macro riverpod.override_with_value}
  Override overrideWithValue(Duration? value) {
    return $ProviderOverride(
      origin: this,
      providerOverride: $SyncValueProvider<Duration?>(value),
    );
  }
}

String _$toastsHash() => r'80b1957ace031e08884658f1df4bfc058463bb08';

abstract class _$Toasts extends $Notifier<Duration?> {
  Duration? build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref = this.ref as $Ref<Duration?, Duration?>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<Duration?, Duration?>,
              Duration?,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
