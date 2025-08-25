import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/data/providers/account_controller.dart';
import 'package:nordvpn/data/providers/popups_provider.dart';
import 'package:nordvpn/data/repository/daemon_status_codes.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/router/router.dart';
import 'package:nordvpn/widgets/dialog_factory.dart';
import 'package:nordvpn/widgets/popups/decision_popup.dart';
import 'package:nordvpn/widgets/popups/info_popup.dart';
import 'package:nordvpn/widgets/popups/rich_popup.dart';

final class PopupsListener extends ConsumerStatefulWidget {
  final Widget child;

  const PopupsListener({super.key, required this.child});

  @override
  ConsumerState<PopupsListener> createState() => _PopupsListenerState();
}

final class _PopupsListenerState extends ConsumerState<PopupsListener> {
  int? _visiblePopup;

  @override
  Widget build(BuildContext _) {
    ref.listen(accountControllerProvider, (_, next) {
      next.whenData((account) {
        if (account == null) return;
        if (!account.isExpired &&
            _visiblePopup == DaemonStatusCode.accountExpired) {
          closeCurrentPopup();
        }
      });
    });
    final popupMetadata = ref.watch(popupsProvider);
    if (popupMetadata != null) {
      _showNextPopup(popupMetadata);
    }
    return widget.child;
  }

  void closeCurrentPopup() {
    final ctx = goRouterKey.currentContext;
    if (ctx == null) {
      logger.e("Can't close popup. Context is null.");
      return;
    }
    DialogFactory.close(ctx);
    _visiblePopup = null;
  }

  void _showNextPopup(PopupMetadata metadata) async {
    final ctx = goRouterKey.currentContext;
    if (ctx == null) {
      logger.e("Can't display popup. Context is null.");
      return;
    }

    _visiblePopup = metadata.id;

    await showDialog(context: ctx, builder: (_) => buildPopup(metadata));

    ref.read(popupsProvider.notifier).pop();
    _visiblePopup = null;
  }
}

Widget buildPopup(PopupMetadata metadata) {
  switch (metadata) {
    case RichPopupMetadata metadata:
      return RichNotificationPopup(metadata: metadata);
    case DecisionPopupMetadata metadata:
      return DecisionPopup(metadata: metadata);
    case InfoPopupMetadata metadata:
      return InfoPopup(metadata: metadata);
  }
}
