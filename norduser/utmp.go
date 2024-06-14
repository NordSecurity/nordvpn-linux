package norduser

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <utmpx.h>

const int ERROR_REALLOC = -1;
const int ERROR_MALLOC_USERNAME = -2;

void free_users_table(char*** users, int size) {
  for (int index = 0; index < size; index++) {
    if ((*users)[index] != NULL) {
      free((*users)[index]);
      (*users)[index] = NULL;
    }
  }

  if (*users != NULL) {
    free(*users);
    *users = NULL;
  }
}

int get_utmp_user_processes(char*** users) {
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
      char** tmp;
      tmp = realloc(*users, (size + 1) * sizeof(char*));
      if (tmp == NULL) {
        free_users_table(users, index);
        endutxent();
        return ERROR_REALLOC;
      }
      *users = tmp;

      size++;
    }

    (*users)[index] = malloc(__UT_NAMESIZE + 1);
    if ((*users)[index] == NULL) {
      free_users_table(users, size);
      endutxent();
      return ERROR_MALLOC_USERNAME;
    }
    (*users)[index][__UT_NAMESIZE]='\0';

    strncpy((*users)[index], u->ut_user, __UT_NAMESIZE);
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

func getActiveUsers() ([]string, error) {
	var usersCArray **C.char
	size := C.get_utmp_user_processes(&usersCArray)

	if size == C.ERROR_REALLOC {
		return []string{}, fmt.Errorf("failed to reallocate space for the users table")
	} else if size == C.ERROR_MALLOC_USERNAME {
		return []string{}, fmt.Errorf("failed to allocate space for new user in the users table")
	}

	log.Printf("%s %d active user processes found", internal.DebugPrefix, size)

	users := []string{}
	for i := 0; i < int(size); i++ {
		usernameCStr := (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(usersCArray)) + uintptr(i)*unsafe.Sizeof(*usersCArray)))
		username := C.GoString(*usernameCStr)
		users = append(users, username)
	}

	C.free_users_table(&usersCArray, size)
	return users, nil
}
