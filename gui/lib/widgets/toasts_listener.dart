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
    final path = ref.watch(currentRoutePathProvider).path;
    final isBlocking = routeRegistry[path]?.isBlocking ?? true;
    final theme = context.toastTheme;

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
  }
}
