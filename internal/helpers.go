package internal

import "time"

// Will try to send the value non-blocking. Possible loosing value.
// if there is no receiver and the channel buffer is full or unbuffered, then the value is lost
func SendNonBlocking[T any](c chan<- T, value T) {
	select {
	case c <- value:
	case <-time.After(500 * time.Millisecond):
	}
}
