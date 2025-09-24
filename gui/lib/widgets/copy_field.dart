import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:nordvpn/i18n/strings.g.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/theme/copy_field_theme.dart';
import 'package:nordvpn/widgets/dynamic_theme_image.dart';

final class CopyItem {
  final String command;
  final String? description;
  const CopyItem({required this.command, this.description});
}

final class CopyField extends StatelessWidget {
  final List<CopyItem> items;

  const CopyField({super.key, required this.items});

  @override
  Widget build(BuildContext context) {
    final appTheme = context.appTheme;

    return LayoutBuilder(
      builder: (context, constraints) {
        return Padding(
          padding: EdgeInsets.only(bottom: appTheme.verticalSpaceLarge),
          child: FractionallySizedBox(
            widthFactor: 0.5,
            child: Column(
              spacing: appTheme.verticalSpaceMedium,
              children: [
                for (final item in items) _buildCopyItem(context, item),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildCopyItem(BuildContext context, CopyItem item) {
    final appTheme = context.appTheme;
    final copyFieldTheme = context.copyFieldTheme;

    return Column(
      spacing: appTheme.verticalSpaceSmall,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (item.description != null)
          Text(item.description!, style: copyFieldTheme.descriptionTextStyle),
        Container(
          decoration: BoxDecoration(
            color: appTheme.areaBackgroundColor,
            borderRadius: BorderRadius.circular(copyFieldTheme.borderRadius),
          ),
          child: Row(
            children: [
              Expanded(
                child: _textArea(appTheme, copyFieldTheme, item.command),
              ),
              _copyButton(item.command),
            ],
          ),
        ),
      ],
    );
  }

  Widget _textArea(
    AppTheme appTheme,
    CopyFieldTheme copyFieldTheme,
    String text,
  ) {
    final lines = max('\n'.allMatches(text).length + 1, 1);
    // 4 lines is a safe value to expand considering our app min sizes
    final maxLines = min(4, lines);
    return Scrollbar(
      child: SingleChildScrollView(
        child: Padding(
          padding: EdgeInsets.all(appTheme.margin),
          child: SelectableText(
            text.trim(),
            style: copyFieldTheme.commandTextStyle,
            textAlign: TextAlign.left,
            minLines: 1,
            maxLines: maxLines,
          ),
        ),
      ),
    );
  }

  IconButton _copyButton(String text) {
    return IconButton(
      tooltip: t.ui.copy,
      onPressed: () {
        Clipboard.setData(ClipboardData(text: text));
      },
      icon: DynamicThemeImage("copy_icon.svg"),
    );
  }
}
