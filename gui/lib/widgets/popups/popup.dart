import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nordvpn/data/models/popup_metadata.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/internal/scaler_responsive_box.dart';
import 'package:nordvpn/theme/popup_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

// Base class providing "template" for popups.
abstract class Popup extends ConsumerWidget {
  final PopupMetadata metadata;

  const Popup({super.key, required this.metadata});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final screenSize = MediaQuery.sizeOf(context);
    final popupTheme = context.popupTheme;
    final theme = Theme.of(context);

    return _AnnounceOnShow(
      announcement: t.a11y.popupWithContent(
        content: "$title. ${message(ref)}",
      ),
      child: Dialog(
        backgroundColor: Colors.transparent,
        child: Container(
          decoration: BoxDecoration(
            color: theme.colorScheme.surface,
            borderRadius: popupTheme.widgetRadius,
          ),
          padding: EdgeInsets.all(popupTheme.verticalElementSpacing),
          width: min(
            dynamicScale(popupTheme.widgetWidth),
            screenSize.width * 0.8,
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            mainAxisSize: MainAxisSize.min,
            children: [
              _titleBar(context, popupTheme),
              buildContent(context, ref),
            ],
          ),
        ),
      ),
    );
  }

  Widget _titleBar(BuildContext context, PopupTheme theme) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Expanded(
          child: Row(
            spacing: theme.contentAllPadding,
            children: [
              if (leadingIcon != null) leadingIcon!,
              Flexible(child: _title(theme)),
            ],
          ),
        ),
        _closeIcon(context),
      ],
    );
  }

  // the title and the message are kept in their own semantics nodes, otherwise
  // they are merged into the dialog and read again for each of its buttons
  Widget _title(PopupTheme theme) {
    return Semantics(
      container: true,
      child: Text(title, style: theme.textPrimary),
    );
  }

  Widget _closeIcon(BuildContext context) {
    final theme = context.popupTheme;
    return IconButton(
      padding: EdgeInsetsGeometry.all(theme.xButtonAllPadding),
      icon: Semantics(
        label: t.ui.close,
        child: DynamicThemeImage("close.svg"),
      ),
      onPressed: () => Navigator.of(context).pop(),
    );
  }

  void closePopup(BuildContext context) => Navigator.of(context).pop();
  String get title => metadata.title ?? t.ui.nordVpn;
  String message(WidgetRef ref) => metadata.message(ref);

  Widget? get leadingIcon => null;
  Widget buildContent(BuildContext context, WidgetRef ref);
}

final class _AnnounceOnShow extends StatefulWidget {
  final String announcement;
  final Widget child;

  const _AnnounceOnShow({required this.announcement, required this.child});

  @override
  State<_AnnounceOnShow> createState() => _AnnounceOnShowState();
}

final class _AnnounceOnShowState extends State<_AnnounceOnShow> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _announce());
  }

  void _announce() {
    if (!mounted || !MediaQuery.supportsAnnounceOf(context)) {
      return;
    }

    SemanticsService.sendAnnouncement(
      View.of(context),
      widget.announcement,
      Directionality.of(context),
    );
  }

  @override
  Widget build(BuildContext context) => widget.child;
}
