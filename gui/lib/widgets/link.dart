import 'package:flutter/material.dart';
import 'package:nordvpn/internal/uri_launch_extension.dart';
import 'package:nordvpn/theme/app_theme.dart';

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
    final theme = context.appTheme;
    return TextButton(
      style: TextButton.styleFrom(padding: EdgeInsets.zero),
      onPressed: () async => await launchableUri.launch(),
      child: Text(
        title,
        style: size == LinkSize.normal ? theme.linkNormal : theme.linkSmall,
        overflow: TextOverflow.ellipsis,
      ),
    );
  }
}

// Convert URI to LaunchableUri class to have launch method
LaunchableUri _toLaunchable<T>(T uri) {
  if (uri is LaunchableUri) return uri;
  if (uri is Uri) return LaunchableUri(uri);
  throw ArgumentError('Unsupported type $T: $uri');
}
