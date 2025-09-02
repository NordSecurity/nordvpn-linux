import 'package:go_router/go_router.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

// Description for a stateful shell branch.
// Information stored into this will help to create the navigation trails
class CustomStatefulShellBranch extends StatefulShellBranch {
  final DynamicThemeImage icon;
  final String label;

  CustomStatefulShellBranch({
    required super.routes,
    required this.icon,
    required this.label,
  });
}
