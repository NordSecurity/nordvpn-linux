import 'package:flutter/material.dart';

/// A custom implementation of ListTile with additional configuration options.
///
/// This widget extends the functionality of the standard ListTile by adding:
/// * Configurable minimum tile height through [minTileHeight]
/// * Enable/disable functionality through [enabled] parameter
/// * Configurable content padding through [contentPadding]
///
/// The widget maintains all standard ListTile features including:
/// * [leading] widget displayed at the start
/// * [title] primary content
/// * [subtitle] secondary content
/// * [trailing] widget displayed at the end
/// * [onTap] callback for tap interactions
///
/// The tile is automatically set to dense layout for compact presentation.
///
/// Example:
/// ```dart
/// CustomListTile(
///   leading: Icon(Icons.star),
///   title: Text('Title'),
///   subtitle: Text('Subtitle'),
///   minTileHeight: 60,
///   onTap: () => print('Tapped'),
/// )
/// ```
class CustomListTile extends StatelessWidget {
  final Widget? leading;
  final Widget? title;
  final Widget? subtitle;
  final Widget? trailing;
  final VoidCallback? onTap;
  final double? minTileHeight;
  final bool enabled;
  final EdgeInsetsGeometry? contentPadding;

  const CustomListTile({
    super.key,
    this.leading,
    this.title,
    this.subtitle,
    this.trailing,
    this.onTap,
    this.minTileHeight,
    this.enabled = true,
    this.contentPadding,
  });

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: leading,
      title: title,
      subtitle: subtitle,
      trailing: trailing,
      onTap: enabled ? onTap : null,
      minTileHeight: minTileHeight,
      contentPadding: contentPadding,
    );
  }
}
