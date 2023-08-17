//@Auther   : Kaishin
//@Time     : 2022/12/22

package internal

import "time"

const (
	autoSyncInterval   = time.Minute
	dialTimeout        = 5 * time.Second
	dialKeepAliveTime  = 5 * time.Second
	endpointsSeparator = ","
)
