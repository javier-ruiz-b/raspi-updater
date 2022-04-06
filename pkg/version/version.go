package version

import "fmt"

const MAJOR = 1
const MINOR = 0
const PATCH = 0

var VERSION = fmt.Sprint(MAJOR, ".", MINOR, ".", PATCH)
