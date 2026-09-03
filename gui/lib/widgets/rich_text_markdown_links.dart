import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/theme/support_link_theme.dart';
import 'package:url_launcher/url_launcher.dart';

// Custom class implementation for RichText that replaces the URL link from
// [label](<url>) to a clickable url
class RichTextMarkdownLinks extends StatefulWidget {
  final String text;
  final TextStyle? style;
  // One callback per link, in the order the links appear in [text].
  // Each is invoked when its link is tapped, and before the URL is launched.
  // The URL is launched regardless of the callback.
  // Number of callbacks (when provided) must match the number of links in [text].
  final List<VoidCallback>? onLinkTaps;

  const RichTextMarkdownLinks({
    super.key,
    required this.text,
    this.style,
    this.onLinkTaps,
  });

  @override
  State<RichTextMarkdownLinks> createState() => _RichTextMarkdownLinksState();
}

// Matches [label](url)
final _linkPattern = RegExp(
  r'\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)',
  caseSensitive: false,
);

// drop the URL, if present in text arg, and keep only the label for the sake of
// screen reader announcements clarity
String dropURLLinkFromPopupMessage(String text) => text.replaceAllMapped(
  _linkPattern,
  (match) => t.a11y.linkWithinPopup(name: match.group(1)!),
);

class _RichTextMarkdownLinksState extends State<RichTextMarkdownLinks> {
  // keep a list with all the TapGestureRecognizer because they need to be disposed manually
  final List<TapGestureRecognizer> _tapGestureRecognizers = [];

  @override
  void dispose() {
    for (final tap in _tapGestureRecognizers) {
      tap.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final linkTheme = context.supportLinkTheme;

    return RichText(
      textScaler: MediaQuery.textScalerOf(context),
      text: TextSpan(
        style: widget.style ?? linkTheme.textStyle,
        children: _buildSpans(context),
      ),
    );
  }

  List<TextSpan> _buildSpans(BuildContext context) {
    final linkTheme = context.supportLinkTheme;
    List<TextSpan> spans = [];
    final matches = _linkPattern.allMatches(widget.text).toList();
    final onLinkTaps = widget.onLinkTaps;
    assert(
      onLinkTaps == null || onLinkTaps.length == matches.length,
      "There are ${matches.length} link(s) but ${onLinkTaps.length} onLinkTap callbacks: ${widget.text}",
    );
    int lastMatchEnd = 0;

    for (final (index, match) in matches.indexed) {
      // Add text before the match
      if (match.start > lastMatchEnd) {
        spans.add(
          TextSpan(text: widget.text.substring(lastMatchEnd, match.start)),
        );
      }

      final label = match.group(1)!;
      final url = match.group(2)!;

      final tap = TapGestureRecognizer()
        ..onTap = () async {
          onLinkTaps?.elementAtOrNull(index)?.call();
          final uri = Uri.parse(url);
          if (!await canLaunchUrl(uri)) {
            logger.e("failed to launch $uri");
          }
          await launchUrl(uri, mode: LaunchMode.externalApplication);
        };

      _tapGestureRecognizers.add(tap);

      spans.add(
        TextSpan(
          text: label,
          style: TextStyle(color: linkTheme.urlColor),
          recognizer: tap,
        ),
      );

      lastMatchEnd = match.end;
    }

    // Add remaining text
    if (lastMatchEnd < widget.text.length) {
      spans.add(TextSpan(text: widget.text.substring(lastMatchEnd)));
    }

    return spans;
  }
}
