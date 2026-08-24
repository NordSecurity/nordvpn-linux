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
  static const semanticsKey = Key("popup_semantics");
  static const titleKey = Key("popup_title");
  static const messageKey = Key("popup_message");
  static const closeButtonKey = Key("popup_close_button");

  final PopupMetadata metadata;

  const Popup({super.key, required this.metadata});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final screenSize = MediaQuery.sizeOf(context);
    final popupTheme = context.popupTheme;
    final theme = Theme.of(context);

    return _AnnounceOnShow(
      announcement: () => semanticLabel(ref),
      child: Semantics(
        key: semanticsKey,
        scopesRoute: true,
        explicitChildNodes: true,
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

  // `MergeSemantics` is required, otherwise `header` is set on this node while
  // the label stays on the child text and the flag has no effect.
  Widget _title(PopupTheme theme) {
    return MergeSemantics(
      child: Semantics(
        header: true,
        child: Text(title, key: titleKey, style: theme.textPrimary),
      ),
    );
  }

  Widget _closeIcon(BuildContext context) {
    final theme = context.popupTheme;
    // `tooltip` covers pointer and keyboard users, the `Semantics` label names
    // the icon for the screen reader. Same pairing as `input.dart` and
    // `bin_button.dart`; the button role comes from IconButton itself.
    return IconButton(
      key: closeButtonKey,
      padding: EdgeInsetsGeometry.all(theme.xButtonAllPadding),
      tooltip: t.ui.close,
      icon: Semantics(label: t.ui.close, child: DynamicThemeImage("close.svg")),
      onPressed: () => Navigator.of(context).pop(),
    );
  }

  void closePopup(BuildContext context) => Navigator.of(context).pop();
  String get title => metadata.title ?? t.ui.nordVpn;
  String message(WidgetRef ref) => metadata.message(ref);

  // Accessible name of the popup, read by the screen reader when it opens.
  // Subclasses override it when their visible heading is not [title].
  @protected
  String semanticLabel(WidgetRef ref) => joinSemanticLabel(title, message(ref));

  @protected
  String joinSemanticLabel(String heading, String body) => body.isEmpty
      ? heading
      : t.a11y.popupWithContent(title: heading, message: body);

  Widget? get leadingIcon => null;
  Widget buildContent(BuildContext context, WidgetRef ref);
}

// Announces the popup once, when it appears. This is imperative on purpose: a
// name on the dialog node would be re-read on every focus change inside the
// popup, prepending the title and the message to each of its buttons.
final class _AnnounceOnShow extends StatefulWidget {
  final ValueGetter<String> announcement;
  final Widget child;

  const _AnnounceOnShow({required this.announcement, required this.child});

  @override
  State<_AnnounceOnShow> createState() => _AnnounceOnShowState();
}

final class _AnnounceOnShowState extends State<_AnnounceOnShow> {
  Animation<double>? _transition;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback(
      (_) => _scheduleAnnouncement(),
    );
  }

  // Wait for the popup to finish appearing: announcing while the route is still
  // animating competes with the speech for the route change itself.
  void _scheduleAnnouncement() {
    if (!mounted) return;

    // No route to wait for when a popup is built directly, as in tests.
    final transition = ModalRoute.of(context)?.animation;
    if (transition == null || transition.isCompleted) {
      _announce();
      return;
    }

    _transition = transition..addStatusListener(_onTransitionStatus);
  }

  void _onTransitionStatus(AnimationStatus status) {
    if (status != AnimationStatus.completed) return;

    _stopListening();
    _announce();
  }

  void _stopListening() {
    _transition?.removeStatusListener(_onTransitionStatus);
    _transition = null;
  }

  void _announce() {
    if (!mounted || !MediaQuery.supportsAnnounceOf(context)) return;

    SemanticsService.sendAnnouncement(
      View.of(context),
      widget.announcement(),
      Directionality.of(context),
    );
  }

  @override
  void dispose() {
    _stopListening();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => widget.child;
}
