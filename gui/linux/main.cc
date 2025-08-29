#include "nordvpn.h"

int main(int argc, char** argv) {
  gtk_init(&argc, &argv);
  
  g_autoptr(NordVPNApplication) app = nordvpn_application_new();
  return g_application_run(G_APPLICATION(app), argc, argv);
}
