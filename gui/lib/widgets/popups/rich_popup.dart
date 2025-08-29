import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/popups/popup.dart';

// Popup with styled header, message, one button and image on the right.
final class RichNotificationPopup extends Popup {
  final RichPopupMetadata richMetadata;

  const RichNotificationPopup({super.key, required super.metadata})
    : richMetadata = metadata as RichPopupMetadata;

  @override
  Widget buildContent(BuildContext context, WidgetRef ref) {
    final theme = context.appTheme;
    return Padding(
      padding: EdgeInsets.only(
        left: theme.outerPadding,
        right: theme.outerPadding,
        bottom: theme.outerPadding,
      ),
      child: Row(
        children: [
          Expanded(
            flex: 5,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              spacing: theme.verticalSpaceMedium,
              children: [
                _header(theme),
                _message(ref, theme),
                SizedBox(
                  width: double.infinity,
                  child: _actionButton(context, ref),
                ),
              ],
            ),
          ),
          const Expanded(flex: 1, child: SizedBox.shrink()),
          Expanded(flex: 2, child: richMetadata.image),
        ],
      ),
    );
  }

  Widget _header(AppTheme theme) {
    return Text(richMetadata.header, style: theme.title);
  }

  Widget _message(WidgetRef ref, AppTheme theme) {
    return Text(richMetadata.message(ref), style: theme.body);
  }

  Widget _actionButton(BuildContext context, WidgetRef ref) {
    return ElevatedButton(
      onPressed: () async {
        await richMetadata.action(ref);
        if (richMetadata.autoClose) {
          if (!context.mounted) return;
          closePopup(context);
        }
      },
      child: Text(richMetadata.actionButtonText),
    );
  }

  @override
  Widget get leadingIcon {
    return DynamicThemeImage("nordvpn_logo.svg");
  }
}
