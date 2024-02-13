import 'package:flutter/material.dart';
import 'package:nordvpn/theme.dart';

class AppHeader extends StatefulWidget implements PreferredSizeWidget {
  AppHeader({super.key});

  @override
  State<AppHeader> createState() => _AppHeaderState();

  @override
  Size get preferredSize => Size.fromHeight(ThemeManager.appHeaderHeight);
}

class _AppHeaderState extends State<AppHeader> {
  bool _isLoggedIn = false;
  bool _isConnected = false;

  @override
  Widget build(BuildContext context) {
    return AppBar(
      toolbarHeight: ThemeManager.appHeaderHeight,
      backgroundColor: ThemeManager.appHeaderBgColor,
      title: SizedBox(
        height: ThemeManager.appHeaderHeight - 20,
        child: Row(children: [
          const Spacer(),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 50),
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(ThemeManager.borderRadius),
              border: Border.all(
                color: ThemeManager.borderColor,
              ),
            ),
            child: Row(
              children: [
                Row(
                  children: [
                    Icon(
                      _isConnected ? Icons.lock : Icons.lock_open,
                      color: _isConnected ? Colors.green : Colors.red,
                    ),
                    const SizedBox(width: 16),
                    Text(
                      "NOT CONNECTED",
                      style: TextStyle(
                        color: _isConnected ? Colors.green : Colors.red,
                      ),
                    ),
                  ],
                ),
                VerticalDivider(
                    color: ThemeManager.borderColor, endIndent: 5, indent: 5),
                Text(_isConnected
                    ? "100.100.0.1"
                    : "Connect now or pick a country"),
              ],
            ),
          ),
          Expanded(
            child: Container(),
          )
        ]),
      ),
      actions: [
        if (!_isLoggedIn)
          Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: ElevatedButton(
                onPressed: () {},
                style: ElevatedButton.styleFrom(
                  backgroundColor: ThemeManager.accentColor,
                  foregroundColor: ThemeManager.accentTextColor,
                  padding:
                      const EdgeInsets.symmetric(vertical: 16, horizontal: 24),
                  shape: RoundedRectangleBorder(
                    borderRadius:
                        BorderRadius.circular(ThemeManager.borderRadius),
                  ),
                ),
                child: const Text('Register'),
              )),
        Padding(
            padding: const EdgeInsets.only(right: 40.0),
            child: ElevatedButton(
              onPressed: () {
                setState(() {
                  if (!_isLoggedIn) {
                    _isLoggedIn = true;
                  } else {
                    _isConnected = !_isConnected;
                  }
                });
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: ThemeManager.accentColor,
                foregroundColor: ThemeManager.accentTextColor,
                padding:
                    const EdgeInsets.symmetric(vertical: 16, horizontal: 24),
                shape: RoundedRectangleBorder(
                  borderRadius:
                      BorderRadius.circular(ThemeManager.borderRadius),
                ),
              ),
              child: Text(!_isLoggedIn
                  ? "Sign in"
                  : (!_isConnected ? "Quick connect" : "Disconnect")),
            )),
      ],
    );
  }
}
