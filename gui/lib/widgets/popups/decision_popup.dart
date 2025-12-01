import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/theme/popup_theme.dart';
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
            Expanded(child: _yesButton(ref, context)),
          ],
        ),
      ],
    );
  }

  Widget _noButton(BuildContext context) {
    return OutlinedButton(
      onPressed: () => closePopup(context),
      child: Text(decisionMetadata.noButtonText),
    );
  }

  Widget _yesButton(WidgetRef ref, BuildContext context) {
    return ElevatedButton(
      onPressed: () {
        closePopup(context);
        decisionMetadata.yesAction(ref);
      },
      child: Text(decisionMetadata.yesButtonText),
    );
  }
}
