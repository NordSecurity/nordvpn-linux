import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/theme/support_link_theme.dart';
import 'package:url_launcher/url_launcher.dart';

// Custom class implementation for RichText that replaces the URL link from
// [label](<url>) to a clickable url
class RichTextMarkdownLinks extends StatefulWidget {
  final String text;
  final TextStyle? style;

  const RichTextMarkdownLinks({super.key, required this.text, this.style});

  @override
  State<RichTextMarkdownLinks> createState() => _RichTextMarkdownLinksState();
}

class _RichTextMarkdownLinksState extends State<RichTextMarkdownLinks> {
  // Matches [label](url)
  final _linkPattern = RegExp(
    r'\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)',
    caseSensitive: false,
  );

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
    final matches = _linkPattern.allMatches(widget.text);
    int lastMatchEnd = 0;

    for (final match in matches) {
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
