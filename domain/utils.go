package domain

import "github.com/sirupsen/logrus"

func CheckError(err error) {
	if err != nil {
		logrus.Errorf("err=%+#v", err)
		panic(err)
	}
}
