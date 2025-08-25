#ifndef FLUTTER_NORDVPN_APPLICATION_H_
#define FLUTTER_NORDVPN_APPLICATION_H_

#include <gtk/gtk.h>

G_DECLARE_FINAL_TYPE(NordVPNApplication, nordvpn_application, NORDVPN, APPLICATION,
                     GtkApplication)

NordVPNApplication* nordvpn_application_new();

#endif  // FLUTTER_NORDVPN_APPLICATION_H_
