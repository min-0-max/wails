//go:build darwin && !ios && !server

package mainthreadharness

/*
#include <pthread/qos.h>
*/
import "C"

// QualityOfServiceUserInteractive is the class of the main thread of an
// application.
const QualityOfServiceUserInteractive = int(C.QOS_CLASS_USER_INTERACTIVE)

// QualityOfServiceOfThisThread returns the quality of service class of the
// calling thread.
func QualityOfServiceOfThisThread() int { return int(C.qos_class_self()) }
