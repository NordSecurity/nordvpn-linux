import 'package:flutter/material.dart';
import 'package:nordvpn/internal/delayed_loading_manager.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/theme/app_theme.dart';
import 'package:nordvpn/widgets/enabled_widget.dart';
import 'package:nordvpn/widgets/inline_loading_indicator.dart';

final class LoadingCheckbox extends StatefulWidget {
  final bool value;
  final String text;
  final Future<void> Function(bool)? onChanged;

  const LoadingCheckbox({
    super.key,
    required this.value,
    required this.text,
    required this.onChanged,
  });

  @override
  State<LoadingCheckbox> createState() => _LoadingCheckboxState();
}

final class _LoadingCheckboxState extends State<LoadingCheckbox> {
  bool _value = false;
  late DelayedLoadingManager _loadingManager;

  @override
  void initState() {
    super.initState();
    _value = widget.value;
    _loadingManager = DelayedLoadingManager(
      onUpdate: _rebuild,
      onDone: () => setState(() => _value = !_value),
    );
  }

  void _rebuild() {
    if (mounted) setState(() {});
  }

  @override
  Widget build(BuildContext context) {
    final theme = context.appTheme;
    return GestureDetector(
      onTap: widget.onChanged == null ? null : _toggle,
      child: Row(
        children: [
          Stack(
            alignment: Alignment.center,
            children: [
              if (_loadingManager.isLoading) InlineLoadingIndicator(),
              EnabledWidget(
                enabled: !_loadingManager.isLoading,
                child: Checkbox(
                  value: _value,
                  onChanged: widget.onChanged == null ? null : (_) => _toggle(),
                ),
              ),
            ],
          ),
          Expanded(
            child: Text(
              widget.text,
              style: widget.onChanged == null ? theme.textDisabled : theme.body,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _toggle() async {
    assert(widget.onChanged != null);
    _loadingManager.startLoading();
    widget.onChanged!
        .call(!_value)
        .then((_) {
          _loadingManager.stopLoading(false);
        })
        .catchError((e) {
          logger.e("error while toggling checkbox: $e");
          _loadingManager.stopLoading(true);
        });
  }

  @override
  void dispose() {
    _loadingManager.dispose();
    super.dispose();
  }
}
