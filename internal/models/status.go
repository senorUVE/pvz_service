package models

import "fmt"

type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusClose      Status = "close"
)

func (Status) Parse(str string) (Status, error) {
	switch str {
	case string(StatusInProgress):
		return StatusInProgress, nil
	case string(StatusClose):
		return StatusClose, nil
	}
	return "", fmt.Errorf("invalid role: %s", str)

}

func (s Status) String() string {
	return string(s)
}
