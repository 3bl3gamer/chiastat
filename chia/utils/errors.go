package utils

import "github.com/ansel1/merry"

func WrongRespError(obj FromBytes) error {
	return merry.Errorf("unexpected response: %#v", obj).WithStackSkipping(1)
}
