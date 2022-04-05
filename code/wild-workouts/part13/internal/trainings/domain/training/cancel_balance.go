package training

import "fmt"

// CancelBalanceDelta возвращает дельту изменения баланса тренировок, который должен быть изменен после отмены тренировки.
func CancelBalanceDelta(tr Training, cancelingUserType UserType) int {
	if tr.CanBeCanceledForFree() {
		// просто возвращаем кредит потраченный на тренировку
		return 1
	}

	switch cancelingUserType {
	case Trainer:
		// 1 за отмену тренировки, + 1 - пеня за отмену тренером менее, чем за 24 часа до тренировки
		return 2
	case Attendee:
		// пеня за отмену тренировки менее, чем за 24 часа
		return 0
	default:
		panic(fmt.Sprintf("not supported user type %s", cancelingUserType))
	}
}
