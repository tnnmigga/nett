package msgbus

import (
	"time"

	"github.com/tnnmigga/nett/conf"
	"github.com/tnnmigga/nett/idef"
	"github.com/tnnmigga/nett/util"
)

type castOpt struct {
	key   string
	value any
}

func findCastOpt[T any](opts []castOpt, key string, defaultVal T) (value T) {
	for _, opt := range opts {
		if opt.key == key {
			return opt.value.(T)
		}
	}
	return defaultVal
}

func NonuseStream() castOpt {
	return castOpt{
		key:   idef.ConstKeyNonuseStream,
		value: true,
	}
}

func OneOfMods(modName idef.ModName) castOpt {
	return castOpt{
		key:   idef.ConstKeyOneOfMods,
		value: modName,
	}
}

func ServerID(serverID uint32) castOpt {
	return castOpt{
		key:   idef.ConstKeyServerID,
		value: serverID,
	}
}

func ServerType(serverType string) castOpt {
	return castOpt{
		key:   idef.ConstKeyServerType,
		value: serverType,
	}
}

func Expires(expires time.Duration) castOpt {
	return castOpt{
		key:   idef.ConstKeyExpires,
		value: int64(util.NowNs() + expires),
	}
}

func Local() castOpt {
	return castOpt{
		key:   idef.ConstKeyServerID,
		value: conf.ServerID,
	}
}
