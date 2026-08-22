package utils

import (
	"context"
	"strconv"
	"time"

	"github.com/mcstatus-io/mcutil/v4/status"
)

func GetStatusVersion(port string) string {
	ip := "127.0.0.1"

	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	response, err := status.Modern(ctx, ip, uint16(portNum))
	if err != nil {
		return ""
	}

	versionName := response.Version.Name.Clean

	return versionName
}
