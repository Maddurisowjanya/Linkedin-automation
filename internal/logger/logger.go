package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// New returns a preconfigured Logrus logger.
// In a larger system you might want structured fields or hooks; this keeps it
// intentionally simple but production‑friendly.
func New() *logrus.Logger {
	log := logrus.New()
	log.Out = os.Stdout

	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// Text formatter is easier to read in terminals; JSON would also be valid.
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return log
}



