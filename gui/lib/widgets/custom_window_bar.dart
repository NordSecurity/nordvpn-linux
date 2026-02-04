
import 'package:flutter/material.dart';
import 'package:nordvpn/i18n/strings.g.dart' show t;
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:window_manager/window_manager.dart';

class CustomTitleBar extends StatefulWidget {
  const CustomTitleBar({super.key});

  @override
  State<CustomTitleBar> createState() => _CustomTitleBarState();
}

class _CustomTitleBarState extends State<CustomTitleBar> {
  bool _isMaximized = false;

  @override
  void initState() {
    super.initState();
    _loadWindowState();
  }

  Future<void> _loadWindowState() async {
    _isMaximized = await windowManager.isMaximized();
    if (mounted) setState(() {});
  }

  Future<void> _toggleMaximize() async {
    if (await windowManager.isMaximized()) {
      await windowManager.restore();
      _isMaximized = false;
    } else {
      await windowManager.maximize();
      _isMaximized = true;
    }

    if (mounted) setState(() {});
  }

  Future<void> _doNothing() async {

  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      // Drag window
      onPanStart: (_) => windowManager.startDragging(),

      // Double click to maximize
      onDoubleTap: _toggleMaximize,

      child: Container(
        height: 45,
        color: Colors.grey.shade900,
        padding: const EdgeInsets.only(top: 2, bottom: 3, left: 20, right: 0),
        child: Row(
          children: [
            const Spacer(),

                      // Insert the search bar
            const SizedBox(
              width: 400,
              child: WindowSearchBar(),
            ),
              

            _WindowButton(icon: Icons.notifications, onPressed: _doNothing),

            _WindowButton(icon: Icons.account_box, onPressed: _doNothing),

            _WindowButton(
              icon: Icons.minimize,
              onPressed: () => windowManager.minimize(),
            ),

            _WindowButton(
              icon: _isMaximized
                  ? Icons.filter_none
                  : Icons.crop_square,
              onPressed: _toggleMaximize,
            ),

            _WindowButton(
              icon: Icons.close,
              hoverColor: Colors.red,
              onPressed: () => windowManager.close(),
            ),
          ],
        ),
      ),
    );
  }
}

class _WindowButton extends StatelessWidget {
  final IconData icon;
  final VoidCallback onPressed;
  final Color? hoverColor;

  const _WindowButton({
    required this.icon,
    required this.onPressed,
    this.hoverColor,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onPressed,
      hoverColor: hoverColor ?? Colors.grey.shade800,
      child: SizedBox(
        width: 46,
        height: double.infinity,
        child: Icon(icon, size: 18),
      ),
    );
  }
}

class WindowSearchBar extends StatelessWidget {
  final double height;

  const WindowSearchBar({this.height = 32, super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      height: height,
      padding: const EdgeInsets.symmetric(horizontal: 8),
      decoration: BoxDecoration(
        color: Colors.grey.shade800,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Row(
        children: [
          // Magnifying glass icon
          Icon(Icons.search, size: 18, color: Colors.white70),

          const SizedBox(width: 6),

          // Placeholder text
          const Expanded(
            child: Text(
              'Search all locations',
              style: TextStyle(color: Colors.white54),
            ),
          ),
        ],
      ),
    );
  }
}