import 'package:nordvpn/i18n/strings.g.dart';

// Matches [label](url)
final _linkPattern = RegExp(
  r'\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)',
  caseSensitive: false,
);

sealed class MarkdownPart {
  const MarkdownPart();
}

final class MarkdownPlainText extends MarkdownPart {
  final String text;

  const MarkdownPlainText(this.text);
}

final class MarkdownLink extends MarkdownPart {
  final String label;
  final String url;

  const MarkdownLink({required this.label, required this.url});
}

// Holds both raw and parsed markdown for further reuse
final class MarkdownText {
  final String raw;
  final List<MarkdownPart> parts;

  MarkdownText.parse(String text) : raw = text, parts = _parse(text);

  int get linksCount => parts.whereType<MarkdownLink>().length;

  // The text as a screen reader should announce it: links keep only their
  // label, dropping the `[label](url)` markdown syntax and the URL itself.
  String get semanticsText => parts
      .map(
        (part) => switch (part) {
          MarkdownPlainText(:final text) => text,
          MarkdownLink(:final label) => t.a11y.linkWithinPopup(name: label),
        },
      )
      .join();
}

List<MarkdownPart> _parse(String text) {
  final parts = <MarkdownPart>[];
  var lastMatchEnd = 0;

  for (final match in _linkPattern.allMatches(text)) {
    // Add text before the match
    if (match.start > lastMatchEnd) {
      parts.add(MarkdownPlainText(text.substring(lastMatchEnd, match.start)));
    }

    parts.add(MarkdownLink(label: match.group(1)!, url: match.group(2)!));
    lastMatchEnd = match.end;
  }

  // Add remaining text
  if (lastMatchEnd < text.length) {
    parts.add(MarkdownPlainText(text.substring(lastMatchEnd)));
  }

  return parts;
}
