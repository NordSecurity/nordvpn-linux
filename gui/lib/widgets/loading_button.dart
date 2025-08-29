import 'dart:async';

import 'package:flutter/material.dart';
import 'package:nordvpn/internal/delayed_loading_manager.dart';
import 'package:nordvpn/logger.dart';
import 'package:nordvpn/widgets/enabled_widget.dart';
import 'package:nordvpn/widgets/inline_loading_indicator.dart';

typedef LoadingTextButton = LoadingButton<TextButton>;
typedef LoadingIconButton = LoadingButton<IconButton>;
typedef LoadingOutlinedButton = LoadingButton<OutlinedButton>;
typedef LoadingElevatedButton = LoadingButton<ElevatedButton>;

enum DisplayModeOnLoading { loadingIndicatorOnly, both }

// Class that will display a loading indicator while the button onPressed
// is running
class LoadingButton<T extends Widget> extends StatefulWidget {
  final FutureOr<void> Function()? onPressed;
  final Widget child;
  final LoadingButtonController? controller;

  // When loading indicator needs to be displayed, specify how to behave:
  // hide the button and have only the loading indicator or display both
  final DisplayModeOnLoading displayModeOnLoading;

  const LoadingButton({
    super.key,
    this.onPressed,
    required this.child,
    this.controller,
    this.displayModeOnLoading = DisplayModeOnLoading.loadingIndicatorOnly,
  });

  @override
  State<LoadingButton> createState() => _LoadingButtonState<T>();
}

// Used to trigger a tap programmatically
final class LoadingButtonController {
  VoidCallback? _onTap;
  void Function(bool)? _onEnableChanged;

  void triggerTap() {
    if (_onTap != null) {
      _onTap!();
    }
  }

  void disable() {
    if (_onEnableChanged != null) {
      _onEnableChanged!(false);
    }
  }

  void enable() {
    if (_onEnableChanged != null) {
      _onEnableChanged!(true);
    }
  }

  void dispose() {
    _onTap = null;
    _onEnableChanged = null;
  }
}

final class _LoadingButtonState<T> extends State<LoadingButton> {
  late DelayedLoadingManager _loadingManager;
  bool _isLoading = false;
  bool _isEnabled = true;

  @override
  void initState() {
    super.initState();
    widget.controller?._onTap = _handleTap;
    widget.controller?._onEnableChanged = (enabled) {
      if (mounted) {
        setState(() {
          _isEnabled = enabled;
        });
      }
    };

    _loadingManager = DelayedLoadingManager(
      onUpdate: () => _rebuild(true),
      onDone: () => _rebuild(false),
      onError: () => _rebuild(false),
    );
  }

  void _rebuild(bool value) {
    if (mounted) setState(() => _isLoading = value);
  }

  @override
  Widget build(BuildContext context) {
    if (widget.displayModeOnLoading == DisplayModeOnLoading.both) {
      return EnabledWidget(
        enabled: !_isLoading,
        disabledOpacity: 1.0,
        child: _createButton(),
      );
    }

    return Stack(
      alignment: Alignment.center,
      fit: StackFit.passthrough,
      children: [
        if (_isLoading) InlineLoadingIndicator(),
        EnabledWidget(enabled: !_isLoading, child: _createButton()),
      ],
    );
  }

  // the button onPressed is assigned to this and the specified onPressed
  //is called inside. This is used to be able to display the loading indicator
  // until onPressed is finished
  Future<void> _handleTap() async {
    if (widget.onPressed == null) {
      return;
    }
    _loadingManager.startLoading();
    try {
      await widget.onPressed!();
      _loadingManager.stopLoading(false);
    } catch (e) {
      logger.e("error while button pressed: $e");
      _loadingManager.stopLoading(true);
    }
  }

  @override
  void dispose() {
    _loadingManager.dispose();
    super.dispose();
  }

  // create the button type based on the T
  Widget _createButton() {
    switch (T) {
      case const (ElevatedButton):
        return ElevatedButton(
          onPressed: _isEnabled && (widget.onPressed != null)
              ? _handleTap
              : null,
          child: _child(true),
        );
      case const (IconButton):
        return IconButton(
          onPressed: _isEnabled && (widget.onPressed != null)
              ? _handleTap
              : null,
          icon: _child(),
        );
      case const (TextButton):
        return TextButton(
          onPressed: _isEnabled && (widget.onPressed != null)
              ? _handleTap
              : null,
          child: _child(),
        );
      case const (OutlinedButton):
        return OutlinedButton(
          onPressed: _isEnabled && (widget.onPressed != null)
              ? _handleTap
              : null,
          child: _child(),
        );
    }
    throw UnimplementedError("widget $T not supported");
  }

  Widget _child([bool alternativeColor = false]) {
    if (_isLoading &&
        (widget.displayModeOnLoading == DisplayModeOnLoading.both)) {
      return Stack(
        alignment: Alignment.center,
        children: [
          Opacity(opacity: 0.0, child: widget.child),
          InlineLoadingIndicator(useAlternativeColor: alternativeColor),
        ],
      );
    }
    return widget.child;
  }
}
