import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:nordvpn/i18n/strings.g.dart';

// Controls if the real daemon should be used or the mock
// To enable for release build add --dart-define=MOCK_GRPC=true to flutter build
const useMockDaemon = bool.fromEnvironment(
  'MOCK_GRPC',
  defaultValue: kDebugMode,
);

const logFile = "nordvpn-gui.log";
final daemonDateFormat = DateFormat("yyyy-MM-dd h:m:s");

const animationDuration = Duration(milliseconds: 200);
const loginTimeoutDuration = Duration(seconds: 10);

const maxInt32 = 0xffffffff;
const maxInt16 = 0xffff;
const maxCustomDnsServers = 3;
// Maximum number of servers returned when searching after the server number '#'
const maxNumberOfServersResults = 50;

// Specify after how many characters the servers search starts
const serverSearchAfterNumChars = 2;
// used to match a string representing a subnet in the format 0.0.0.0/32
final RegExp subnetFormatPattern = RegExp(
  r'^(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])'
  r'\.(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])'
  r'\.(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])'
  r'\.(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])'
  r'/(3[0-2]|[12]?\d)$',
);
// Match IPv4 for custom DNS
final RegExp ipv4Regex = RegExp(
  r'^(?:(?:25[0-5]|2[0-4][0-9]|1\d{2}|[1-9]?\d)(?:\.(?!$)|$)){4}$',
);

const defaultTheme = ThemeMode.system;

// Main window sizes
final windowMinSize = Size(460, 574);
final windowDefaultSize = Size(800, 600);
final windowMaxSize = Size.fromWidth(1200);
final fastestServerLabel = "${t.ui.fastestServer} (${t.ui.quickConnect})";

// server group backend names
const doubleVpn = "Double_vpn";
const dedicatedIp = "Dedicated_IP";
const onionOverVpn = "Onion_Over_VPN";
const p2p = "p2p";
const obfuscatedServers = "Obfuscated_Servers";
