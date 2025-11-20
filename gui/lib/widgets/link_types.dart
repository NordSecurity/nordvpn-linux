import 'package:nordvpn/widgets/link.dart';

/// A clickable link for first-party websites.
///
/// Use for links to pages within the company's ecosystem.
final class FirstPartyLink<T> extends Link<T> {
  FirstPartyLink({super.key, required super.title, required super.uri});
}
