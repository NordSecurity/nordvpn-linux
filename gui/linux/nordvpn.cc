#include "nordvpn.h"

#include <flutter_linux/flutter_linux.h>
#include <glib-unix.h>
#ifdef GDK_WINDOWING_X11
#include <gdk/gdkx.h>
#endif

#include "flutter/generated_plugin_registrant.h"

// Tray popup: pointer to the window so the SIGUSR1 handler can reach it.
static GtkWindow* gTrayWindow = nullptr;

// Positions the tray popup window in the corner near the panel.
// Must be called on X11; no-op on other backends.
static void tray_window_reposition(GtkWindow* window) {
#ifdef GDK_WINDOWING_X11
  GdkDisplay* gdk_display = gdk_display_get_default();
  if (!GDK_IS_X11_DISPLAY(gdk_display)) return;

  GdkMonitor* monitor = gdk_display_get_primary_monitor(gdk_display);
  if (monitor == nullptr) monitor = gdk_display_get_monitor(gdk_display, 0);
  if (monitor == nullptr) return;

  GdkRectangle geometry;
  GdkRectangle workarea;
  gdk_monitor_get_geometry(monitor, &geometry);
  gdk_monitor_get_workarea(monitor, &workarea);

  const int window_width  = 380;
  const int window_height = 640;
  const int margin = 8;

  int x = (workarea.x + workarea.width) - window_width - margin;

  gboolean panel_at_bottom =
      (workarea.y + workarea.height) < (geometry.y + geometry.height);
  int y = panel_at_bottom
      ? (workarea.y + workarea.height) - window_height - margin
      : workarea.y + margin;

  gtk_window_move(window, x, y);
#endif
}

// Called on the GLib main loop when SIGUSR1 is received. Toggles visibility.
static gboolean on_tray_toggle(gpointer) {
  if (gTrayWindow == nullptr) return G_SOURCE_CONTINUE;
  if (gtk_widget_is_visible(GTK_WIDGET(gTrayWindow))) {
    gtk_widget_hide(GTK_WIDGET(gTrayWindow));
  } else {
    tray_window_reposition(gTrayWindow);
    gtk_widget_show(GTK_WIDGET(gTrayWindow));
    gtk_window_present(gTrayWindow);
  }
  return G_SOURCE_CONTINUE;
}

// Called ~150 ms after focus-out: hide only if the window still lacks focus.
// The delay absorbs transient focus changes (e.g. GNOME's AppIndicator briefly
// stealing focus while it opens/closes its empty menu popup on icon click).
static gboolean on_focus_out_delayed(GtkWidget* widget) {
  if (!gtk_window_has_toplevel_focus(GTK_WINDOW(widget))) {
    gtk_widget_hide(widget);
  }
  return G_SOURCE_REMOVE;
}

static gboolean on_focus_out(GtkWidget* widget, GdkEventFocus*, gpointer) {
  g_timeout_add(150, (GSourceFunc)on_focus_out_delayed, widget);
  return FALSE;
}

struct _NordVPNApplication {
  GtkApplication parent_instance;
  char** dart_entrypoint_arguments;
};

G_DEFINE_TYPE(NordVPNApplication, nordvpn_application, GTK_TYPE_APPLICATION)

// Implements GApplication::activate.
static void nordvpn_application_activate(GApplication* application) {
  NordVPNApplication* self = NORDVPN_APPLICATION(application);
  GtkWindow* window =
      GTK_WINDOW(gtk_application_window_new(GTK_APPLICATION(application)));

  // Use a header bar when running in GNOME as this is the common style used
  // by applications and is the setup most users will be using (e.g. Ubuntu
  // desktop).
  // If running on X and not using GNOME then just use a traditional title bar
  // in case the window manager does more exotic layout, e.g. tiling.
  // If running on Wayland assume the header bar will work (may need changing
  // if future cases occur).
  gboolean use_header_bar = TRUE;
#ifdef GDK_WINDOWING_X11
  GdkScreen* screen = gtk_window_get_screen(window);
  if (GDK_IS_X11_SCREEN(screen)) {
    const gchar* wm_name = gdk_x11_screen_get_window_manager_name(screen);
    if (g_strcmp0(wm_name, "GNOME Shell") != 0) {
      use_header_bar = FALSE;
    }
  }
#endif
  if (use_header_bar) {
    GtkHeaderBar* header_bar = GTK_HEADER_BAR(gtk_header_bar_new());
    gtk_widget_show(GTK_WIDGET(header_bar));
    gtk_header_bar_set_title(header_bar, "NordVPN");
    gtk_header_bar_set_show_close_button(header_bar, TRUE);
    gtk_window_set_titlebar(window, GTK_WIDGET(header_bar));
  } else {
    gtk_window_set_title(window, "NordVPN");
  }

  // The minimum size is replaced in flutter code, but it is not always working.
  // Skip in tray mode: the window is 380x640 and Flutter sets its own minimum.
  if (g_getenv("NORDVPN_TRAY_LAUNCH") == nullptr) {
    GdkGeometry hints;
    hints.min_width = 900;
    hints.min_height = 700;
    gtk_window_set_geometry_hints(window, NULL, &hints, GDK_HINT_MIN_SIZE);
  }

  if (g_getenv("NORDVPN_TRAY_LAUNCH") != nullptr) {
    // No title bar or borders.
    gtk_window_set_decorated(window, FALSE);
    // Popup-menu hint: WM will not allow moving or resizing, and the window
    // won't appear in the taskbar/pager.
    gtk_window_set_type_hint(window, GDK_WINDOW_TYPE_HINT_POPUP_MENU);
    // The popup-menu hint disables keyboard focus by default; re-enable it so
    // focus-out events fire and Flutter input works.
    gtk_window_set_accept_focus(window, TRUE);
    // Keep the popup above all other windows.
    gtk_window_set_keep_above(window, TRUE);
    // Hide when the user clicks outside (focus moves to another window).
    g_signal_connect(window, "focus-out-event", G_CALLBACK(on_focus_out), nullptr);
    // SIGUSR1 toggles show/hide; g_unix_signal_add routes it safely through
    // the GLib main loop so we can call GTK functions in the callback.
    gTrayWindow = window;
    g_unix_signal_add(SIGUSR1, on_tray_toggle, nullptr);
  }

  gtk_widget_realize(GTK_WIDGET(window));

  // Position before showing to avoid a flash at (0,0).
  // gtk_window_move only takes effect on X11/XWayland, which is the only
  // backend used in tray mode.
  if (g_getenv("NORDVPN_TRAY_LAUNCH") != nullptr) {
    tray_window_reposition(window);
  }

  if (g_getenv("NORDVPN_TRAY_LAUNCH") != nullptr) {
    // Show immediately; the window is already positioned above.
    gtk_widget_show(GTK_WIDGET(window));
    gtk_window_present(window);
  } else {
    // Normal mode: start hidden, Flutter calls windowManager.show() when ready.
    gtk_widget_hide(GTK_WIDGET(window));
  }

  g_autoptr(FlDartProject) project = fl_dart_project_new();
  fl_dart_project_set_dart_entrypoint_arguments(project, self->dart_entrypoint_arguments);

  FlView* view = fl_view_new(project);
  gtk_widget_show(GTK_WIDGET(view));
  gtk_container_add(GTK_CONTAINER(window), GTK_WIDGET(view));

  fl_register_plugins(FL_PLUGIN_REGISTRY(view));

  gtk_widget_grab_focus(GTK_WIDGET(view));
}

// Implements GApplication::local_command_line.
static gboolean nordvpn_application_local_command_line(GApplication* application, gchar*** arguments, int* exit_status) {
  NordVPNApplication* self = NORDVPN_APPLICATION(application);
  // Strip out the first argument as it is the binary name.
  self->dart_entrypoint_arguments = g_strdupv(*arguments + 1);

  g_autoptr(GError) error = nullptr;
  if (!g_application_register(application, nullptr, &error)) {
     g_warning("Failed to register: %s", error->message);
     *exit_status = 1;
     return TRUE;
  }

  g_application_activate(application);
  *exit_status = 0;

  return TRUE;
}

// Implements GObject::dispose.
static void nordvpn_application_dispose(GObject* object) {
  NordVPNApplication* self = NORDVPN_APPLICATION(object);
  g_clear_pointer(&self->dart_entrypoint_arguments, g_strfreev);
  G_OBJECT_CLASS(nordvpn_application_parent_class)->dispose(object);
}

static void nordvpn_application_class_init(NordVPNApplicationClass* klass) {
  G_APPLICATION_CLASS(klass)->activate = nordvpn_application_activate;
  G_APPLICATION_CLASS(klass)->local_command_line = nordvpn_application_local_command_line;
  G_OBJECT_CLASS(klass)->dispose = nordvpn_application_dispose;
}

static void nordvpn_application_init(NordVPNApplication* self) {}

NordVPNApplication* nordvpn_application_new() {
  return NORDVPN_APPLICATION(g_object_new(nordvpn_application_get_type(),
                                         "application-id", APPLICATION_ID,
                                         "flags", G_APPLICATION_NON_UNIQUE,
                                         nullptr));
}
