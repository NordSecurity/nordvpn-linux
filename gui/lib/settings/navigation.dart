import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nordvpn/router/routes.dart';
import 'package:nordvpn/theme/settings_theme.dart';

final class NavBreadcrumbs extends StatelessWidget {
  const NavBreadcrumbs({super.key});

  @override
  Widget build(BuildContext context) {
    final settingsTheme = context.settingsTheme;
    final location = GoRouter.of(
      context,
    ).routerDelegate.currentConfiguration.uri.toString();
    final uri = Uri.parse(location);
    final segments = uri.pathSegments;
    if (segments.isEmpty ||
        (segments.length == 1 &&
            routeToNameMap[segments[0]]?.displayName == null)) {
      return SizedBox.shrink();
    }

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        children: [
          ConstrainedBox(
            constraints: BoxConstraints(
              minWidth: MediaQuery.sizeOf(context).width,
            ),
            child: Row(
              children: segments.asMap().entries.expand((entry) {
                final idx = entry.key;
                final segment = entry.value;
                final routeMetadata = routeToNameMap[segment];
                assert(
                  routeMetadata?.displayName != null,
                  "missing display name for $entry with $segments",
                );
                if (routeMetadata?.displayName == null) {
                  return [SizedBox.shrink()];
                }
                final displayName = routeMetadata!.displayName!;

                // last breadcrumb is the current page, so no navigation here
                if (idx == segments.length - 1) {
                  return [Breadcrumb(name: displayName)];
                }

                return [
                  NavigableBreadcrumb(
                    name: routeMetadata.displayName!,
                    onPressed: () {
                      if (routeMetadata.onPressed != null) {
                        routeMetadata.onPressed!(context);
                      } else {
                        context.navigateToRoute(routeMetadata.route);
                      }
                    },
                  ),
                  Text(" / ", style: settingsTheme.parentPageStyle),
                ];
              }).toList(),
            ),
          ),
        ],
      ),
    );
  }
}

final class NavigableBreadcrumb extends StatelessWidget {
  final String name;
  final VoidCallback? onPressed;

  const NavigableBreadcrumb({
    super.key,
    required this.name,
    required this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    final theme = context.settingsTheme;
    return TextButton(
      onPressed: onPressed,
      style: TextButton.styleFrom(
        minimumSize: Size.zero,
        padding: EdgeInsets.zero,
        tapTargetSize: MaterialTapTargetSize.shrinkWrap,
      ),
      child: Text(
        name,
        style: theme.parentPageStyle,
        overflow: TextOverflow.ellipsis,
      ),
    );
  }
}

final class Breadcrumb extends StatelessWidget {
  final String name;

  const Breadcrumb({super.key, required this.name});

  @override
  Widget build(BuildContext context) {
    final theme = context.settingsTheme;
    return Text(
      name,
      style: theme.currentPageNameStyle,
      overflow: TextOverflow.ellipsis,
    );
  }
}
