import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/theme/popup_theme.dart';
import 'package:nordvpn/widgets/popups/popup.dart';
import 'package:nordvpn/widgets/rich_text_markdown_links.dart';

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
        Semantics(
          container: true,
          child: RichTextMarkdownLinks(
            text: message(ref), 
            key: Popup.messageKey,
            style: theme.textSecondary,
            onLinkTaps: _linkCallbacks(ref),
          ),
        ),
        Align(alignment: Alignment.centerRight, child: _closeButton(context)),
      ],
    );
  }

  List<VoidCallback>? _linkCallbacks(WidgetRef ref) {
    final onLinkTaps = infoMetadata.onLinkTaps;
    if (onLinkTaps == null) return null;

    final callbacks = <VoidCallback>[];
    for (final onLinkTap in onLinkTaps) {
      callbacks.add(() => onLinkTap(ref));
    }

    return callbacks;
  }

  Widget _closeButton(BuildContext context) {
    final theme = context.popupTheme;
    return ConstrainedBox(
      constraints: BoxConstraints(minWidth: theme.singleButtonMinWidth),
      child: ElevatedButton(
        onPressed: () => closePopup(context),
        child: Text(infoMetadata.buttonText ?? t.ui.close),
      ),
    );
  }
}
