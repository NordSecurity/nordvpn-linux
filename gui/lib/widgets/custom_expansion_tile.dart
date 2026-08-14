import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:nordvpn/i18n/strings.g.dart';

// This is a custom implementation for ExpansionTile, because by default
// clicking on the ExpansionTile would expand/collapse the subitems.
final class CustomExpansionTile extends StatefulWidget {
  final Widget? leading;
  final Widget title;
  final Widget? subtitle;
  final double? minTileHeight;
  final EdgeInsetsGeometry? childrenPadding;
  final List<Widget>? children;
  final VoidCallback? onTap;
  final bool expanded;
  final bool hideExpandButton;
  final bool enabled;
  final Widget? trailing;
  final EdgeInsetsGeometry? contentPadding;

  const CustomExpansionTile({
    super.key,
    required this.title,
    this.children,
    this.subtitle,
    this.onTap,
    this.leading,
    this.minTileHeight,
    this.childrenPadding,
    this.expanded = false,
    this.hideExpandButton = false,
    this.enabled = true,
    this.trailing,
    this.contentPadding,
  });

  @override
  State<StatefulWidget> createState() => _CustomExpansionTileState();
}

class _CustomExpansionTileState extends State<CustomExpansionTile> {
  bool _isExpanded = false;

  @override
  void initState() {
    _isExpanded = widget.expanded;

    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return Material(
      type: MaterialType.transparency,
      child: Column(
        children: [
          ListTile(
            enabled: widget.enabled,
            minTileHeight: widget.minTileHeight,
            contentPadding: widget.contentPadding,
            leading: widget.leading,
            title: widget.title,
            subtitle: widget.subtitle,
            trailing: _buildTrailing(),
            onTap: widget.enabled ? widget.onTap : null,
          ),
          if (_isExpanded && (widget.children != null))
            Padding(
              padding: widget.childrenPadding ?? EdgeInsets.zero,
              child: Column(children: widget.children!),
            ),
        ],
      ),
    );
  }

  Widget? _buildTrailing() {
    if (widget.hideExpandButton || !widget.enabled) {
      return null;
    }

    if (widget.trailing != null) {
      return widget.trailing;
    }

    if (widget.children == null) {
      return null;
    }
    return MergeSemantics(
      child: Semantics(
        label: _isExpanded
            ? t.a11y.expandibleEntryCollapse
            : t.a11y.expandibleEntryExpand,
        expanded: _isExpanded,
        child: IconButton(
          icon: Icon(_isExpanded ? Icons.expand_less : Icons.expand_more),
          onPressed: _toggleExpanded,
          hoverColor: Colors.transparent,
          splashColor: Colors.transparent,
          highlightColor: Colors.transparent,
        ),
      ),
    );
  }

  void _toggleExpanded() {
    setState(() {
      _isExpanded = !_isExpanded;
    });

    if (!MediaQuery.supportsAnnounceOf(context)) {
      return;
    }

    SemanticsService.sendAnnouncement(
      View.of(context),
      _isExpanded
          ? t.a11y.expandibleEntryExpanded
          : t.a11y.expandibleEntryCollapsed,
      Directionality.of(context),
    );
  }
}
