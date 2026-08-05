import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/providers/popup_actions_provider.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/loading_indicator.dart';

// Displays a progress indicator while a popup action is running. The popup
// itself is already closed at that moment, so this is the only feedback the user
// gets until the action is finished.
//
// The indicator blocks the input to prevent starting another change while the
// previous one is still being applied. The delay before displaying it and the
// minimum time it stays visible are handled by [PopupActions], because
// PopupsListener uses the same state to postpone the next popup.
final class PopupActionProgressOverlay extends ConsumerWidget {
  final Widget child;

  const PopupActionProgressOverlay({super.key, required this.child});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final progress = ref.watch(popupActionsProvider);
    final appTheme = context.appTheme;

    return Stack(
      children: [
        child,
        if (progress != null && progress.showIndicator) ...[
          ModalBarrier(
            color: appTheme.overlayBackgroundColor,
            dismissible: false,
          ),
          Center(child: LoadingIndicator(message: progress.message)),
        ],
      ],
    );
  }
}
