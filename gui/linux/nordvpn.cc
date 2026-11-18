#include "nordvpn.h"

#include <flutter_linux/flutter_linux.h>
#ifdef GDK_WINDOWING_X11
#include <gdk/gdkx.h>
#endif

#include "flutter/generated_plugin_registrant.h"

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
  // As a backup set also here to a minimum value.
  GdkGeometry hints;
  hints.min_width = 900;
  hints.min_height = 700;
  gtk_window_set_geometry_hints(window, NULL, &hints, GDK_HINT_MIN_SIZE);

  // Keep in sync with windowDefaultSize in lib/constants.dart. Set before the
  // view is realized, so flutter's first frame is full size, not 1x1 and then
  // resized when X11 maps the window.
  gtk_window_set_default_size(window, 900, 700);

  g_autoptr(FlDartProject) project = fl_dart_project_new();
  fl_dart_project_set_dart_entrypoint_arguments(project, self->dart_entrypoint_arguments);
  // Impeller became the default renderer on Linux in Flutter 3.47 but still
  // has rendering issues. Canonical disables it in Ubuntu's Flutter apps too:
  // https://github.com/canonical/ubuntu-desktop-provision/pull/1618
  fl_dart_project_set_enable_impeller(project, FALSE);

  FlView* view = fl_view_new(project);
  gtk_widget_show(GTK_WIDGET(view));
  gtk_container_add(GTK_CONTAINER(window), GTK_WIDGET(view));
  // Realized but not shown, because window_manager shows the window from Dart.
  gtk_widget_realize(GTK_WIDGET(view));

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
