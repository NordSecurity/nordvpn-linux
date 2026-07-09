import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:nordvpn/theme/context_menu_theme.dart';
import 'package:nordvpn/widgets/context_menu/context_menu_item.dart';

export 'package:nordvpn/widgets/context_menu/context_menu_item.dart';

/// A drop-over context menu anchored to an [anchorBuilder] widget.
///
/// Opens below the anchor when tapped. Tapping outside the menu closes it
/// without propagating the tap to elements underneath.
///
/// Keyboard behavior:
/// - Opening the menu moves focus to the first item.
/// - Tab / Shift+Tab move between items.
/// - Escape closes the menu and returns focus to the anchor.
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

  /// Sizes the menu panel to match the anchor widget's rendered width.
  ///
  /// When true, [width] is ignored.
  final bool matchAnchorWidth;

  const ContextMenu({
    super.key,
    required this.anchorBuilder,
    required this.items,
    this.width,
    this.matchAnchorWidth = false,
  }) : assert(items.length > 0, 'ContextMenu must have at least one item');

  @override
  State<ContextMenu> createState() => _ContextMenuState();
}

class _ContextMenuState extends State<ContextMenu>
    with SingleTickerProviderStateMixin {
  final _layerLink = LayerLink();
  final _overlayController = OverlayPortalController();

  // Owns focus for the open menu panel, so Escape can be scoped to it.
  final _menuScopeNode = FocusScopeNode(debugLabel: 'ContextMenuPanel');

  FocusNode? _previousFocus;

  late final AnimationController _animationController;
  late final Animation<double> _animation;
  bool _initialized = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (!_initialized) {
      _initialized = true;
      final theme = context.contextMenuTheme;
      _animationController = AnimationController(
        vsync: this,
        duration: theme.animationDuration,
      );
      _animation = CurvedAnimation(
        parent: _animationController,
        curve: theme.animationCurve,
      );
    }
  }

  @override
  void dispose() {
    _animationController.dispose();
    _menuScopeNode.dispose();
    super.dispose();
  }

  void _open() {
    _previousFocus = FocusManager.instance.primaryFocus;
    _overlayController.show();
    _animationController.forward();
  }

  void _close() {
    _animationController.reverse().then((_) {
      if (mounted) {
        _overlayController.hide();
        _previousFocus?.requestFocus();
        _previousFocus = null;
      }
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
    _animationController.reverse().then((_) {
      if (mounted) {
        _overlayController.hide();
        itemOnTap();
      }
    });
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
    final menuWidth = widget.matchAnchorWidth
        ? ((this.context.findRenderObject() as RenderBox?)?.size.width ??
              theme.menuWidth)
        : (widget.width ?? theme.menuWidth);

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
          offset: Offset(0, theme.menuGap),
          child: Align(
            alignment: AlignmentDirectional.topStart,
            child: FadeTransition(
              opacity: _animation,
              // Shadow around the menu panel. Needs to be a parent of SizeTransition to avoid clipping.
              child: SizedBox(
                width: menuWidth,
                child: DecoratedBox(
                  decoration: BoxDecoration(
                    borderRadius: theme.menuRadius,
                    border: Border.all(
                      color: theme.menuBorderColor,
                      width: theme.menuBorderWidth,
                    ),
                    boxShadow: theme.menuBoxShadow,
                  ),
                  child: SizeTransition(
                    sizeFactor: _animation,
                    axisAlignment: -1,
                    child: FocusScope(
                      node: _menuScopeNode,
                      child: Shortcuts(
                        shortcuts: const <ShortcutActivator, Intent>{
                          SingleActivator(LogicalKeyboardKey.escape):
                              DismissIntent(),
                        },
                        child: Actions(
                          actions: <Type, Action<Intent>>{
                            DismissIntent: CallbackAction<DismissIntent>(
                              onInvoke: (intent) {
                                _close();
                                return null;
                              },
                            ),
                          },
                          child: FocusTraversalGroup(
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

    return SizedBox(
      width: width,
      child: Material(
        color: theme.menuColor,
        borderRadius: theme.menuRadius,
        clipBehavior: Clip.antiAlias,
        child: Padding(
          padding: theme.menuPadding,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              for (var i = 0; i < items.length; i++)
                _MenuItemTile(
                  item: items[i],
                  onTapped: onItemTapped,
                  autofocus: i == 0,
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _MenuItemTile extends StatelessWidget {
  final ContextMenuItem item;
  final void Function(VoidCallback) onTapped;
  final bool autofocus;

  const _MenuItemTile({
    required this.item,
    required this.onTapped,
    this.autofocus = false,
  });

  @override
  Widget build(BuildContext context) {
    final theme = context.contextMenuTheme;
    final labelStyle = theme.itemTextStyle.copyWith(color: item.labelColor);

    return InkWell(
      key: item.key,
      autofocus: autofocus,
      onTap: () => onTapped(item.onTap),
      hoverColor: theme.itemHoverColor,
      borderRadius: theme.itemBorderRadius,
      mouseCursor: SystemMouseCursors.basic,
      child: Semantics(
        label: item.label,
        button: true,
        enabled: true,
        excludeSemantics: true,
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
