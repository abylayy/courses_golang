package utils

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(1, 3) // Rate limit of 1 request per second with a burst of 3 requests

func isSelected(currentSort, optionSort string) bool {
	return currentSort == optionSort
}

func InitLogger(logger *logrus.Logger) {
	logger.SetFormatter(&logrus.JSONFormatter{})

	file, err := os.OpenFile("logfile.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.SetOutput(file)
	} else {
		fmt.Println("Failed to open log file. Logging to stdout.")
	}

	logger.SetLevel(logrus.InfoLevel)
}

func Seq(start, end int) []int {
	s := make([]int, end-start+1)
	for i := range s {
		s[i] = start + i
	}
	return s
}
