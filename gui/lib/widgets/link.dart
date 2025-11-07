import 'package:flutter/material.dart';
import 'package:nordvpn/internal/uri_launch_extension.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/aurora_design.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

enum LinkSize { normal, small }

final class Link<T> extends StatelessWidget {
  final String title;
  final LaunchableUri launchableUri;
  final LinkSize size;

  Link({
    super.key,
    required this.title,
    required T uri,
    this.size = LinkSize.normal,
  }) : launchableUri = _toLaunchable(uri);

  @override
  Widget build(BuildContext context) {
    return TextButton(
      style: TextButton.styleFrom(padding: EdgeInsets.zero),
      onPressed: () async => await launchableUri.launch(),
      child: _buildContent(context),
    );
  }

  Widget _buildContent(BuildContext context) {
    final theme = context.appTheme;
    return Text(
      title,
      style: size == LinkSize.normal ? theme.linkNormal : theme.linkSmall,
      overflow: TextOverflow.ellipsis,
    );
  }
}

// Convert URI to LaunchableUri class to have launch method
LaunchableUri _toLaunchable<T>(T uri) {
  if (uri is LaunchableUri) return uri;
  if (uri is Uri) return LaunchableUri(uri);
  throw ArgumentError('Unsupported type $T: $uri');
}

/// A clickable link widget with a trailing icon.
///
/// Extends [Link] to display a theme-aware icon alongside the link text.
/// The icon is positioned to the right of the text with [AppSpacing.spacing2] spacing.
///
/// Accepts either a [DynamicThemeImage] object via [icon] or a string path
/// via [iconPath]. Exactly one must be provided. Use [Link] directly if no
/// icon is needed.
final class IconLink<T> extends Link<T> {
  final DynamicThemeImage _icon;

  IconLink({
    super.key,
    required super.title,
    required super.uri,
    super.size,
    DynamicThemeImage? icon,
    String? iconPath,
  }) : assert(
         (icon != null) ^ (iconPath != null),
         'Exactly one of icon or iconPath must be provided',
       ),
       _icon = icon ?? DynamicThemeImage(iconPath!);

  @override
  Widget _buildContent(BuildContext context) {
    final appTheme = context.appTheme;
    return Row(
      spacing: appTheme.verticalSpaceSmall,
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [super._buildContent(context), _icon],
    );
  }
}
