package logger

import "github.com/rs/xid"

func genId() string {
	return xid.New().String()
}
