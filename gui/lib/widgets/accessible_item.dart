import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class AccessibleItem extends StatefulWidget {
  const AccessibleItem({
    super.key,
    required this.child,
    required this.onActivate,
    this.label,
    this.enabled = true,
    this.excludeChildSemantics = true,
    this.button = false,
    this.toggled,
    this.checked,
    this.inMutuallyExclusiveGroup = false,
  });

  final Widget child;

  final VoidCallback onActivate;

  final String? label;
  final bool enabled;

  final bool excludeChildSemantics;

  final bool button;
  final bool? toggled;
  final bool? checked;
  final bool inMutuallyExclusiveGroup;

  @override
  State<AccessibleItem> createState() => _AccessibleItemState();
}

class _AccessibleItemState extends State<AccessibleItem> {
  bool _focused = false;

  @override
  Widget build(BuildContext context) {
    final onActivate = widget.enabled ? widget.onActivate : null;

    Widget visual = ExcludeFocus(child: widget.child);
    if (widget.excludeChildSemantics) {
      visual = ExcludeSemantics(child: visual);
    }

    visual = ColoredBox(
      color: _focused ? Theme.of(context).hoverColor : Colors.transparent,
      child: visual,
    );

    return MergeSemantics(
      child: Semantics(
        button: widget.button,
        toggled: widget.toggled,
        checked: widget.checked,
        inMutuallyExclusiveGroup: widget.inMutuallyExclusiveGroup,
        enabled: widget.enabled,
        label: widget.label,
        onTap: onActivate,
        child: FocusableActionDetector(
          enabled: widget.enabled,
          mouseCursor: widget.enabled
              ? SystemMouseCursors.click
              : SystemMouseCursors.basic,
          onShowFocusHighlight: (value) {
            if (value != _focused) setState(() => _focused = value);
          },
          shortcuts: const <ShortcutActivator, Intent>{
            SingleActivator(LogicalKeyboardKey.enter): ActivateIntent(),
            SingleActivator(LogicalKeyboardKey.space): ActivateIntent(),
          },
          actions: <Type, Action<Intent>>{
            ActivateIntent: CallbackAction<ActivateIntent>(
              onInvoke: (_) {
                onActivate?.call();
                return null;
              },
            ),
          },
          child: visual,
        ),
      ),
    );
  }
}
