import 'package:flutter/material.dart';
import 'package:nordvpn/theme/inline_loading_indicator_theme.dart';

// Loading indicator placed inside (or instead of) a widget.
// It informs user that the widget interaction is in progress.
final class InlineLoadingIndicator extends StatelessWidget {
  final bool useAlternativeColor;
  const InlineLoadingIndicator({super.key, this.useAlternativeColor = false});

  @override
  Widget build(BuildContext context) {
    final theme = context.inlineLoadingIndicatorTheme;
    // add center to ignore parent constraints and always have the given size
    return Center(
      child: SizedBox(
        width: theme.width,
        height: theme.height,
        child: CircularProgressIndicator(
          color: useAlternativeColor ? theme.alternativeColor : theme.color,
          strokeWidth: theme.stroke,
        ),
      ),
    );
  }
}
