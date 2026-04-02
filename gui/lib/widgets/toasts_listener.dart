import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/providers/toasts_provider.dart';
import 'package:nordvpn/theme/toast_theme.dart';
import 'package:nordvpn/widgets/toast.dart';

final class ToastsListener extends ConsumerWidget {
  final Widget child;

  const ToastsListener({super.key, required this.child});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    Duration? timeout = ref.watch(toastsProvider);
    final theme = context.toastTheme;
    return Stack(
      children: [
        child,
        if (timeout != null)
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