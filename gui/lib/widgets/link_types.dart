import 'package:nordvpn/widgets/link.dart';

/// A clickable link for first-party websites.
///
/// Use for links to pages within the company's ecosystem.
final class FirstPartyLink<T> extends Link<T> {
  FirstPartyLink({super.key, required super.title, required super.uri});
}

/// A clickable link for third-party websites.
///
/// Use for links to pages outside the company's ecosystem.
final class ThirdPartyLink<T> extends IconLink<T> {
  ThirdPartyLink({super.key, required super.title, required super.uri})
      : super(iconName: "external_link.svg");
}