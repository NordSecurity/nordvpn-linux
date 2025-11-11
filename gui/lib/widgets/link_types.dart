import 'package:nordvpn/widgets/link.dart';

/// A clickable link widget for first-party pages.
///
/// Extends [Link] to provide a simple text link without any icons.
/// Use this for links to publicly accessible pages within the company's ecosystem.
final class InternalLink<T> extends Link<T> {
  InternalLink({super.key, required super.title, required super.uri});
}

/// A clickable link widget for third-party pages.
///
/// Extends [IconLink] to display a link with an external link icon.
/// Use this for links to pages outside the company's ecosystem.
/// The icon is automatically added to indicate an external destination.
final class ExternalLink<T> extends IconLink<T> {
  ExternalLink({super.key, required super.title, required super.uri})
      : super(iconName: "external_link.svg");
}
