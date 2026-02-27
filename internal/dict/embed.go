package dict

import "embed"

//go:embed data/cmudict.txt
var CMUData embed.FS
