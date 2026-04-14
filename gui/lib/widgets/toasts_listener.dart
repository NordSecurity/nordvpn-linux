import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/providers/toasts_provider.dart';
import 'package:nordvpn/router/router.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/theme/toast_theme.dart';
import 'package:nordvpn/widgets/toast.dart';

final class ToastsListener extends ConsumerWidget {
  final Widget child;

  const ToastsListener({super.key, required this.child});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final Duration? timeout = ref.watch(toastsProvider);
    final router = ref.watch(routerProvider);
    final theme = context.toastTheme;

    return ListenableBuilder(
      listenable: router.routerDelegate,
      builder: (context, _) {
        final path = router.routerDelegate.currentConfiguration.uri.path;
        final isBlocking = AppRoute.values.any(
          (r) => r.blocksToast && r.toString() == path,
        );
        return Stack(
          children: [
            child,
            if (timeout != null && !isBlocking)
              Positioned(
                right: theme.widgetPositionRight,
                bottom: theme.widgetPositionBottom,
                child: Toast(
                  duration: timeout,
                  onClose: () => ref.read(toastsProvider.notifier).closeToast(),
                ),
              ),
          ],
        );
      },
    );
  }
}
