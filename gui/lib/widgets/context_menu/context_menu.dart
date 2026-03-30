import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:nordvpn/theme/aurora_design.dart';
import 'package:nordvpn/theme/context_menu_theme.dart';
import 'package:nordvpn/widgets/context_menu/context_menu_item.dart';

export 'package:nordvpn/widgets/context_menu/context_menu_item.dart';

/// A drop-over context menu anchored to an [anchorBuilder] widget.
///
/// Opens below the anchor when tapped. Tapping outside the menu closes it
/// without propagating the tap to elements underneath.
///
/// Example:
/// ```dart
/// ContextMenu(
///   items: [
///     ContextMenuItem(label: 'Option A', onTap: () {}),
///     ContextMenuItem(label: 'Delete', labelColor: Colors.red, onTap: () {}),
///   ],
///   anchorBuilder: (toggle) => OutlinedButton(
///     onPressed: toggle,
///     child: const Text('Open'),
///   ),
/// )
/// ```
final class ContextMenu extends StatefulWidget {
  /// Builds the widget that triggers the menu.
  ///
  /// The [toggleMenu] callback opens the menu when it is closed and closes it
  /// when it is open. Wire it to the anchor widget's tap handler.
  final Widget Function(VoidCallback toggleMenu) anchorBuilder;

  /// The list of items to display in the menu.
  final List<ContextMenuItem> items;

  /// Override the menu panel width. Defaults to [ContextMenuTheme.menuWidth].
  final double? width;

  const ContextMenu({
    super.key,
    required this.anchorBuilder,
    required this.items,
    this.width,
  }) : assert(items.length > 0, 'ContextMenu must have at least one item');

  @override
  State<ContextMenu> createState() => _ContextMenuState();
}

class _ContextMenuState extends State<ContextMenu>
    with SingleTickerProviderStateMixin {
  final _layerLink = LayerLink();
  final _overlayController = OverlayPortalController();
  final _keyboardFocusNode = FocusNode();
  late final AnimationController _animationController;
  late final Animation<double> _fadeAnimation;
  late final Animation<double> _sizeAnimation;

  @override
  void initState() {
    super.initState();
    _animationController = AnimationController(
      vsync: this,
      duration: AppTransitions.durationFast,
    );
    _fadeAnimation = CurvedAnimation(
      parent: _animationController,
      curve: AppTransitions.timingFunctionDefault,
    );
    _sizeAnimation = CurvedAnimation(
      parent: _animationController,
      curve: AppTransitions.timingFunctionDefault,
    );
  }

  @override
  void dispose() {
    _keyboardFocusNode.dispose();
    _animationController.dispose();
    super.dispose();
  }

  void _open() {
    _overlayController.show();
    _animationController.forward();
  }

  void _close() {
    _animationController.reverse().then((_) {
      if (mounted) _overlayController.hide();
    });
  }

  void _toggle() {
    if (_overlayController.isShowing) {
      _close();
    } else {
      _open();
    }
  }

  void _onItemTapped(VoidCallback itemOnTap) {
    _close();
    itemOnTap();
  }

  @override
  Widget build(BuildContext context) {
    return OverlayPortal(
      controller: _overlayController,
      overlayChildBuilder: _buildOverlay,
      child: CompositedTransformTarget(
        link: _layerLink,
        child: widget.anchorBuilder(_toggle),
      ),
    );
  }

  Widget _buildOverlay(BuildContext context) {
    final theme = context.contextMenuTheme;
    final menuWidth = widget.width ?? theme.menuWidth;

    return Stack(
      children: [
        // Barrier: absorbs all taps outside the menu panel.
        Positioned.fill(
          child: GestureDetector(
            behavior: HitTestBehavior.opaque,
            onTap: _close,
            child: const SizedBox.expand(),
          ),
        ),
        // Menu panel anchored to the bottom-left of the anchor widget.
        CompositedTransformFollower(
          link: _layerLink,
          targetAnchor: Alignment.bottomLeft,
          followerAnchor: Alignment.topLeft,
          offset: const Offset(0, 4),
          child: Align(
            alignment: AlignmentDirectional.topStart,
            child: KeyboardListener(
              focusNode: _keyboardFocusNode,
              autofocus: true,
              onKeyEvent: (event) {
                if (event is KeyDownEvent &&
                    event.logicalKey == LogicalKeyboardKey.escape) {
                  _close();
                }
              },
              child: FadeTransition(
                opacity: _fadeAnimation,
                child: SizeTransition(
                  sizeFactor: _sizeAnimation,
                  axisAlignment: -1,
                  child: _MenuPanel(
                    items: widget.items,
                    width: menuWidth,
                    onItemTapped: _onItemTapped,
                  ),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _MenuPanel extends StatelessWidget {
  final List<ContextMenuItem> items;
  final double width;
  final void Function(VoidCallback) onItemTapped;

  const _MenuPanel({
    required this.items,
    required this.width,
    required this.onItemTapped,
  });

  @override
  Widget build(BuildContext context) {
    final theme = context.contextMenuTheme;
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final shadows = isDark ? AppBoxShadows.darkPopover : AppBoxShadows.lightPopover;

    return Container(
      width: width,
      decoration: BoxDecoration(
        borderRadius: theme.menuRadius,
        border: Border.all(
          color: theme.menuBorderColor,
          width: theme.menuBorderWidth,
        ),
        boxShadow: shadows,
      ),
      child: Material(
        color: theme.menuColor,
        borderRadius: theme.menuRadius,
        clipBehavior: Clip.antiAlias,
        child: Padding(
          padding: theme.menuPadding,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: items
                .map((item) => _MenuItemTile(item: item, onTapped: onItemTapped))
                .toList(),
          ),
        ),
      ),
    );
  }
}

class _MenuItemTile extends StatelessWidget {
  final ContextMenuItem item;
  final void Function(VoidCallback) onTapped;

  const _MenuItemTile({required this.item, required this.onTapped});

  @override
  Widget build(BuildContext context) {
    final theme = context.contextMenuTheme;
    final labelStyle = theme.itemTextStyle.copyWith(
      color: item.labelColor,
    );

    return Semantics(
      button: true,
      label: item.label,
      child: InkWell(
        onTap: () => onTapped(item.onTap),
        hoverColor: theme.itemHoverColor,
        child: Padding(
          padding: theme.itemPadding,
          child: SizedBox(
            height: theme.itemHeight,
            child: Align(
              alignment: AlignmentDirectional.centerStart,
              child: Text(item.label, style: labelStyle),
            ),
          ),
        ),
      ),
    );
  }
}
