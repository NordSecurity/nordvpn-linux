import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/theme/popup_theme.dart';
import 'package:nordvpn/widgets/dialog_factory.dart';
import 'package:nordvpn/widgets/popups/popup.dart';

// Popup with title, message and two buttons (yes/no).
final class DecisionPopup extends Popup {
  final DecisionPopupMetadata decisionMetadata;

  const DecisionPopup({super.key, required super.metadata})
    : decisionMetadata = metadata as DecisionPopupMetadata;

  @override
  Widget buildContent(BuildContext context, WidgetRef ref) {
    final theme = context.popupTheme;
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      spacing: theme.verticalElementSpacing,
      children: [
        Text(message(ref), style: theme.textSecondary),
        Row(
          spacing: theme.gapBetweenElements,
          children: [
            Expanded(child: _noButton(context)),
            Expanded(child: _yesButton(context)),
          ],
        ),
      ],
    );
  }

  // The yes/no actions are intentionally not invoked here. This popup only
  // reports the choice, closing itself in the process. PopupsListener then hands
  // the action to `PopupActions`, which runs it with a container scoped ref.
  // Running it here would use this widget's `ref` after the dialog route was
  // disposed and throw "Cannot use ref after the widget was disposed.".
  Widget _noButton(BuildContext context) {
    return OutlinedButton(
      onPressed: () => Navigator.of(context).pop(DialogResult.no),
      child: Text(decisionMetadata.noButtonText),
    );
  }

  Widget _yesButton(BuildContext context) {
    return ElevatedButton(
      onPressed: () => Navigator.of(context).pop(DialogResult.yes),
      child: Text(decisionMetadata.yesButtonText),
    );
  }
}
