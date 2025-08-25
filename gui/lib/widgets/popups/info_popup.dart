import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/popups/popup.dart';

// Popup for showing information (like failed action). It can be only closed.
final class InfoPopup extends Popup {
  final InfoPopupMetadata infoMetadata;

  const InfoPopup({super.key, required super.metadata})
    : infoMetadata = metadata as InfoPopupMetadata;

  @override
  Widget buildContent(BuildContext context, WidgetRef ref) {
    final appTheme = context.appTheme;

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      spacing: appTheme.horizontalSpace,
      children: [
        Text(message(ref), style: appTheme.body),
        Align(alignment: Alignment.centerRight, child: _closeButton(context)),
      ],
    );
  }

  Widget _closeButton(BuildContext context) {
    return ElevatedButton(
      onPressed: () => closePopup(context),
      child: Text(t.ui.close),
    );
  }
}
