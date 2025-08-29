import 'package:flutter/material.dart';
import 'package:nordvpn/theme/interactive_list_view_theme.dart';
import 'package:searchable_listview/searchable_listview.dart';

// A list with a text field to support filtering the items
final class InteractiveListView extends StatefulWidget {
  final String? searchHintText;
  // This will be displayed inside the text field before the text
  final Widget? leadingWidget;

  // the error widget when a search and no results are found
  final Widget noResultsFoundWidget;
  // when not null it will be displayed when nothing is searched instead of empty list
  final Widget? emptyListWidget;
  final bool showEmptyListAtStartup;
  final List<dynamic> items;
  final Widget Function(BuildContext context, dynamic item) itemBuilder;
  final List<dynamic> Function(String query, List<dynamic> items) filter;
  final TextStyle searchBarSize;
  final TextEditingController? searchTextController;
  final int beginSearchAfter;

  const InteractiveListView({
    super.key,
    this.searchHintText,
    this.leadingWidget,
    required this.items,
    required this.itemBuilder,
    required this.filter,
    required this.noResultsFoundWidget,
    required this.searchBarSize,
    required this.showEmptyListAtStartup,
    required this.emptyListWidget,
    this.searchTextController,
    required this.beginSearchAfter,
  });

  @override
  State<InteractiveListView> createState() => _InteractiveListViewState();
}

class _InteractiveListViewState extends State<InteractiveListView> {
  final _searchController = TextEditingController();
  final _searchFieldNodeFocus = FocusNode();

  @override
  void initState() {
    super.initState();
    assert(widget.beginSearchAfter >= 1);
  }

  @override
  void dispose() {
    _searchController.dispose();
    _searchFieldNodeFocus.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final listView = _buildSearchableList(context);
    _searchFieldNodeFocus.requestFocus();
    return listView;
  }

  Widget _buildSearchableList(BuildContext context) {
    var initialItems = widget.items;
    if (widget.showEmptyListAtStartup) {
      initialItems = [];
    }

    final theme = context.interactiveListViewTheme;
    return SearchableList<dynamic>(
      focusNode: _searchFieldNodeFocus,
      searchTextController: widget.searchTextController ?? _searchController,
      displaySearchIcon: true,
      maxLines: 1,
      textStyle: widget.searchBarSize,
      itemBuilder: (item) => widget.itemBuilder(context, item),
      initialList: initialItems,
      filter: (query) {
        if (query.length < widget.beginSearchAfter) {
          // rebuild the list with error
          setState(() {});
          return [];
        }
        final filteredItems = widget.filter(query, widget.items);

        if ((widget.showEmptyListAtStartup) && filteredItems.isEmpty) {
          // rebuild the list with error
          setState(() {});
        }

        return filteredItems;
      },
      emptyWidget: _emptyResultsWidget(),
      inputDecoration: InputDecoration(
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(theme.borderRadius),
          borderSide: BorderSide(
            color: theme.borderColor,
            width: theme.borderWidth,
          ),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(theme.borderRadius),
          borderSide: BorderSide(
            color: theme.focusBorderColor,
            width: theme.borderWidth,
          ),
        ),
        hintText: widget.searchHintText,
        icon: widget.leadingWidget,
      ),
      closeKeyboardWhenScrolling: true,
    );
  }

  Widget _emptyResultsWidget() {
    if ((widget.showEmptyListAtStartup) &&
        (_searchController.text.length < widget.beginSearchAfter)) {
      return widget.emptyListWidget ?? const SizedBox.shrink();
    }
    return widget.noResultsFoundWidget;
  }
}
