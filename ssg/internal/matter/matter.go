package matter

import (
	"errors"
	"time"
)

type Matter struct {
	Name  string    `yaml:"name"`
	Date  time.Time `yaml:"date"`
	Tags  []string  `yaml:"tags"`
	Draft bool      `yaml:"draft"`
}

func (m *Matter) Validate() error {
	if m.Name == "" {
		return errors.New("Mising matter name.")
	}

	if m.Date.IsZero() {
		return errors.New("Date is zero.")
	}

	return nil
}
