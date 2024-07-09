package norduser

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <utmpx.h>

const int ERROR_REALLOC = -1;
const int ERROR_MALLOC_USERNAME = -2;

typedef struct user {
  pid_t login_pid;
  char username[__UT_NAMESIZE + 1];
} user;

int get_utmp_user_processes(user** users) {
  int index = 0;
  int size = 0;
  setutxent();

  *users = NULL;

  struct utmpx* u;
  while ((u = getutxent()) != NULL) {
    if (u->ut_type != USER_PROCESS) {
      continue;
    }

    if (index == size) {
      user* tmp;
      tmp = realloc(*users, (size + 1) * sizeof(user));
      if (tmp == NULL) {
        free(*users);
        endutxent();
        return ERROR_REALLOC;
      }
      *users = tmp;

      size++;
    }

	(*users)[index].login_pid = u->ut_pid;
    strncpy((*users)[index].username, u->ut_user, __UT_NAMESIZE);
	(*users)[index].username[__UT_NAMESIZE] = '\0';
    index++;
  }

  endutxent();

  return size;
}
*/
import "C"
import (
	"fmt"
	"log"
	"unsafe"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

type userData map[string]norduserState

// getActiveUsers returns a map of [username]userType where type can be text or gui. If any of the given users login
// processes will be detected to have a gui(done based on the environment), user type will be gui. Otherwise user type
// will be text.
func getActiveUsers() (userData, error) {
	var usersCArray *C.user
	size := C.get_utmp_user_processes(&usersCArray)
	if size == C.ERROR_REALLOC {
		return userData{}, fmt.Errorf("failed to reallocate space for the users table")
	} else if size == C.ERROR_MALLOC_USERNAME {
		return userData{}, fmt.Errorf("failed to allocate space for new user in the users table")
	}
	defer C.free(unsafe.Pointer(usersCArray))

	log.Printf("%s %d active user processes found", internal.DebugPrefix, size)

	users := make(userData)
	for index := 0; index < int(size); index++ {
		userC := (*C.user)(unsafe.Pointer(uintptr(unsafe.Pointer(usersCArray)) +
			uintptr(index)*unsafe.Sizeof(*usersCArray)))
		usernameC := (*C.char)(unsafe.Pointer(&userC.username))
		username := C.GoString(usernameC)

		// userType was determined to be gui for any of the users login processes, so we skip this user.
		if userType, ok := users[username]; ok && userType == loginGUI {
			continue
		}

		loginPID := uint32(userC.login_pid)

		desktopSession, err := findEnvVariableForPID(loginPID, "XDG_CURRENT_DESKTOP")
		if err != nil {
			log.Printf("%s looking up XDG_CURRENT_DESKTOP for %s: %s", internal.ErrorPrefix, username, err)
			continue
		}

		userType := loginText
		if desktopSession != "" {
			userType = loginGUI
		}

		users[username] = userType
	}

	return users, nil
}
