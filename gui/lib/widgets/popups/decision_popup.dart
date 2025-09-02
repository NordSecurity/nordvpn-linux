import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/popups/popup.dart';

// Popup with title, message and two buttons (yes/no).
final class DecisionPopup extends Popup {
  final DecisionPopupMetadata decisionMetadata;

  const DecisionPopup({super.key, required super.metadata})
    : decisionMetadata = metadata as DecisionPopupMetadata;

  @override
  Widget buildContent(BuildContext context, WidgetRef ref) {
    final appTheme = context.appTheme;

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      spacing: appTheme.horizontalSpace,
      children: [
        Text(message(ref), style: appTheme.body),
        Row(
          spacing: appTheme.verticalSpaceSmall,
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
      child: Text(decisionMetadata.yesButtonText),
      onPressed: () async {
        await decisionMetadata.yesAction(ref);
        if (!context.mounted) return;
        closePopup(context);
      },
    );
  }
}
