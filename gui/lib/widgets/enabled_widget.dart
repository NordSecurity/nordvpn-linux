import 'package:flutter/material.dart';

// Used to control the mouse and focus status for its children widgets
final class EnabledWidget extends StatelessWidget {
  final bool enabled;
  final double disabledOpacity;
  final Widget child;

  const EnabledWidget({
    super.key,
    required this.enabled,
    required this.child,
    this.disabledOpacity = 0.0,
  });

  @override
  Widget build(BuildContext context) {
    return ExcludeFocus(
      excluding: !enabled,
      child: IgnorePointer(
        ignoring: !enabled,
        child: Opacity(opacity: enabled ? 1.0 : disabledOpacity, child: child),
      ),
    );
  }
}
