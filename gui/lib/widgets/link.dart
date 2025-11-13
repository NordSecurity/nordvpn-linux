import 'package:flutter/material.dart';
import 'package:nordvpn/internal/uri_launch_extension.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

enum LinkSize { normal, small }

class Link<T> extends StatelessWidget {
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
      style: TextButton.styleFrom(
        padding: EdgeInsets.symmetric(vertical: 2),
        minimumSize: Size.zero,
        tapTargetSize: MaterialTapTargetSize.shrinkWrap,
      ),
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

/// A clickable link with a trailing icon.
///
/// Use for links that need a visual indicator icon.
class IconLink<T> extends Link<T> {
  final DynamicThemeImage _icon;

  IconLink({
    super.key,
    required super.title,
    required super.uri,
    super.size,
    required String iconName,
  }) : _icon = DynamicThemeImage(iconName);

  @override
  Widget _buildContent(BuildContext context) {
    final appTheme = context.appTheme;
    return Row(
      spacing: appTheme.horizontalSpaceSmall,
      crossAxisAlignment: CrossAxisAlignment.end,
      children: [super._buildContent(context), _icon],
    );
  }
}
