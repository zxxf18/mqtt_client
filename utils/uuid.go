package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func UUID() string {
	u, err := uuid.NewUUID()
	if err != nil {
		fmt.Printf("failed to generate uuid: %s\n", err.Error())
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return u.String()
}
