import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/theme/popup_theme.dart';
import 'package:nordvpn/widgets/popups/popup.dart';

// Popup for showing information (like failed action). It can be only closed.
final class InfoPopup extends Popup {
  final InfoPopupMetadata infoMetadata;

  const InfoPopup({super.key, required super.metadata})
    : infoMetadata = metadata as InfoPopupMetadata;

  @override
  Widget buildContent(BuildContext context, WidgetRef ref) {
    final theme = context.popupTheme;

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      spacing: theme.verticalElementSpacing,
      children: [
        Text(message(ref), style: theme.textSecondary),
        Align(alignment: Alignment.centerRight, child: _closeButton(context)),
      ],
    );
  }

  Widget _closeButton(BuildContext context) {
    final theme = context.popupTheme;
    return ConstrainedBox(
      constraints: BoxConstraints(
        minWidth: theme.singleButtonMinWidth,
        minHeight: theme.buttonHeight,
      ),
      child: ElevatedButton(
        onPressed: () => closePopup(context),
        style: ElevatedButton.styleFrom(
          padding: theme.buttonPadding,
          backgroundColor: theme.primaryButtonBackgroundColor,
        ),
        child: Text(t.ui.close),
      ),
    );
  }
}
