#include <X11/Xlib.h>

#include "nordvpn.h"


int main(int argc, char** argv) {
  // fix for https://github.com/NordSecurity/nordvpn-linux/issues/1136
  // ensure X11 threads are initialized before using any X11 functionality
  XInitThreads(); 
  gtk_init(&argc, &argv);
  
  g_autoptr(NordVPNApplication) app = nordvpn_application_new();
  return g_application_run(G_APPLICATION(app), argc, argv);
}
