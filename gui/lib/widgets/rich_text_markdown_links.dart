import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:nordvpn/internal/markdown_text.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/theme/support_link_theme.dart';
import 'package:url_launcher/url_launcher.dart';

// Custom class implementation for RichText that replaces the URL link from
// [label](<url>) to a clickable url
class RichTextMarkdownLinks extends StatefulWidget {
  final String? text;
  final MarkdownText? markdown;
  final TextStyle? style;
  // One callback per link, in the order the links appear in [text].
  // Each is invoked when its link is tapped, and before the URL is launched.
  // The URL is launched regardless of the callback.
  // Number of callbacks (when provided) must match the number of links in [text].
  final List<VoidCallback>? onLinkTaps;

  const RichTextMarkdownLinks({
    super.key,
    this.text,
    this.markdown,
    this.style,
    this.onLinkTaps,
  }) : assert(
         (text == null) != (markdown == null),
         "provide either text or markdown",
       );

  @override
  State<RichTextMarkdownLinks> createState() => _RichTextMarkdownLinksState();
}

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
    final markdown = widget.markdown ?? MarkdownText.parse(widget.text!);
    final onLinkTaps = widget.onLinkTaps;
    assert(
      onLinkTaps == null || onLinkTaps.length == markdown.linksCount,
      "There are ${markdown.linksCount} link(s) but ${onLinkTaps.length} onLinkTap callbacks: ${markdown.raw}",
    );
    int linkIndex = 0;

    for (final part in markdown.parts) {
      switch (part) {
        case MarkdownPlainText(:final text):
          spans.add(TextSpan(text: text));

        case MarkdownLink(:final label, :final url):
          final index = linkIndex++;
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
      }
    }

    return spans;
  }
}
