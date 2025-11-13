#include "nordvpn.h"
#include <X11/Xlib.h>

int main(int argc, char** argv) {
  XInitThreads();

  gtk_init(&argc, &argv);
  
  g_autoptr(NordVPNApplication) app = nordvpn_application_new();
  return g_application_run(G_APPLICATION(app), argc, argv);
}
