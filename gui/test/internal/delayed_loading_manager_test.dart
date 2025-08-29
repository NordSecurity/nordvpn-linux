import 'package:fake_async/fake_async.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:nordvpn/internal/delayed_loading_manager.dart';

void main() {
  group('DelayedLoadingManager Tests', () {
    late DelayedLoadingManager manager;

    tearDown(() => manager.dispose());

    test('startLoading sets isLoading immediately', () {
      manager = DelayedLoadingManager(
        delayDuration: const Duration(milliseconds: 50),
        minDisplayDuration: const Duration(milliseconds: 500),
        onUpdate: () {},
        onDone: () {},
      );
      expect(manager.isLoading, isFalse);

      manager.startLoading();

      expect(manager.isLoading, isTrue);
      expect(manager.showLoadingIndicator, isFalse);
    });

    test('Loading indicator shows after delayDuration', () {
      fakeAsync((async) {
        bool updated = false;
        manager = DelayedLoadingManager(
          delayDuration: const Duration(milliseconds: 50),
          minDisplayDuration: const Duration(milliseconds: 500),
          onUpdate: () => updated = true,
          onDone: () {},
        );

        manager.startLoading();
        expect(manager.isLoading, isTrue);
        expect(manager.showLoadingIndicator, isFalse);

        async.elapse(const Duration(milliseconds: 49));
        expect(manager.showLoadingIndicator, isFalse);

        async.elapse(const Duration(milliseconds: 1));
        expect(updated, isTrue);
        expect(manager.showLoadingIndicator, isTrue);
      });
    });

    test('stopLoading before delayDuration, indicator never appears', () {
      fakeAsync((async) {
        manager = DelayedLoadingManager(
          delayDuration: const Duration(milliseconds: 50),
          minDisplayDuration: const Duration(milliseconds: 500),
          onUpdate: () {},
          onDone: () {},
        );
        manager.startLoading();
        expect(manager.isLoading, isTrue);
        expect(manager.showLoadingIndicator, isFalse);

        async.elapse(const Duration(milliseconds: 25));
        manager.stopLoading(false);
        expect(manager.isLoading, isFalse);
        expect(manager.showLoadingIndicator, isFalse);

        async.elapse(const Duration(milliseconds: 25));
        expect(manager.showLoadingIndicator, isFalse);
      });
    });

    test('stopLoading after delayDuration but before minDisplayDuration', () {
      fakeAsync((async) {
        manager = DelayedLoadingManager(
          delayDuration: const Duration(milliseconds: 50),
          minDisplayDuration: const Duration(milliseconds: 500),
          onUpdate: () {},
          onDone: () {},
        );
        manager.startLoading();

        async.elapse(const Duration(milliseconds: 50));
        expect(manager.showLoadingIndicator, isTrue);

        manager.stopLoading(false);
        // we started showing loading indicator, so now show it for minDisplayDuration
        expect(manager.isLoading, isTrue);
        expect(manager.showLoadingIndicator, isTrue);

        async.elapse(const Duration(milliseconds: 499));
        expect(manager.isLoading, isTrue);
        expect(manager.showLoadingIndicator, isTrue);

        async.elapse(const Duration(milliseconds: 1));
        expect(manager.isLoading, isFalse);
        expect(manager.showLoadingIndicator, isFalse);
      });
    });

    test('stopLoading after minDisplayDuration', () {
      fakeAsync((async) {
        manager = DelayedLoadingManager(
          delayDuration: const Duration(milliseconds: 50),
          minDisplayDuration: const Duration(milliseconds: 500),
          onUpdate: () {},
          onDone: () {},
        );
        manager.startLoading();

        async.elapse(const Duration(milliseconds: 50));
        expect(manager.showLoadingIndicator, isTrue);

        async.elapse(const Duration(milliseconds: 500));

        manager.stopLoading(false);
        expect(manager.isLoading, isFalse);
        expect(manager.showLoadingIndicator, isFalse);
      });
    });

    test('onDone is called after loader finishes', () {
      fakeAsync((async) {
        bool isDone = false;
        manager = DelayedLoadingManager(
          delayDuration: const Duration(milliseconds: 50),
          minDisplayDuration: const Duration(milliseconds: 500),
          onUpdate: () {},
          onDone: () => isDone = true,
        );
        manager.startLoading();
        expect(isDone, isFalse);

        async.elapse(const Duration(milliseconds: 50));
        expect(isDone, isFalse);
        expect(manager.showLoadingIndicator, isTrue);

        manager.stopLoading(false);
        expect(manager.showLoadingIndicator, isTrue);
        expect(isDone, isFalse); // loading indicator still visible

        async.elapse(const Duration(milliseconds: 500));
        expect(manager.showLoadingIndicator, isFalse);
        expect(isDone, isTrue);
        expect(manager.isLoading, isFalse);
      });
    });

    test('dispose cancels timers and prevents updates', () {
      fakeAsync((async) {
        int updatedCount = 0;
        manager = DelayedLoadingManager(
          delayDuration: const Duration(milliseconds: 50),
          minDisplayDuration: const Duration(milliseconds: 500),
          onUpdate: () {
            updatedCount += 1;
          },
          onDone: () {},
        );
        manager.startLoading();
        manager.dispose();

        async.elapse(const Duration(milliseconds: 100));

        expect(manager.isLoading, isTrue);
        expect(manager.showLoadingIndicator, isFalse);
        expect(updatedCount, equals(1)); // updated only once on startLoading
      });
    });
  });
}
