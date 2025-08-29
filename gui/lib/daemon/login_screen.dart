import 'package:flutter/material.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';
import 'package:nordvpn/widgets/full_screen_scaffold.dart';
import 'package:nordvpn/widgets/login_form.dart';

final class LoginScreen extends StatelessWidget {
  const LoginScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;
    return FullScreenScaffold(
      child: Padding(
        padding: const EdgeInsets.only(left: 40, right: 40),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.start,
          children: [
            const Flexible(child: LoginForm()),
            SizedBox(width: appTheme.verticalSpaceMedium),
            DynamicThemeImage("login_screen_image.svg"),
          ],
        ),
      ),
    );
  }
}
